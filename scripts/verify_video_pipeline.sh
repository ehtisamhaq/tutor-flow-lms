#!/bin/bash
set -e

BASE_URL="http://localhost:8080/api/v1"
ADMIN_EMAIL="admin@tutorflow.com"
ADMIN_PASS="password123"

echo "Verifying Secure Video Pipeline..."

# 1. Login
echo "Logging in..."
LOGIN_RES=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASS\"}")

TOKEN=$(echo $LOGIN_RES | jq -r '.data.tokens.access_token')

if [ "$TOKEN" == "null" ]; then
  echo "Login failed: $LOGIN_RES"
  exit 1
fi
echo "Login successful. Token acquired."

# 2. Create Course (if needed) - For simplicity, let's assume we can upload to a new lesson
# We need a lesson ID. Let's create a dummy course structure.
COURSE_TITLE="Video Test Course $(date +%s)"
echo "Creating dummy course: $COURSE_TITLE"
COURSE_Res=$(curl -s -X POST $BASE_URL/courses \
  -H "Authorization: Bearer $TOKEN" \
  -d "title=$COURSE_TITLE" \
  -d "description=Testing HLS Pipeline" \
  -d "price=0" \
  -d "level=beginner" \
  -d "category_id=development") # Assuming category_id is string or optional? code says FormValue.
COURSE_ID=$(echo $COURSE_Res | jq -r '.data.id')
echo "Course created: $COURSE_ID"

if [ "$COURSE_ID" == "null" ]; then
  echo "Course creation failed: $COURSE_Res"
  exit 1
fi

echo "Creating dummy module..."
MODULE_RES=$(curl -s -X POST $BASE_URL/courses/$COURSE_ID/modules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Module", "sort_order": 1}')
MODULE_ID=$(echo $MODULE_RES | jq -r '.data.id')
echo "Module created: $MODULE_ID"

if [ "$MODULE_ID" == "null" ]; then
  echo "Module creation failed: $MODULE_RES"
  exit 1
fi

echo "Creating dummy lesson..."
LESSON_RES=$(curl -s -X POST $BASE_URL/courses/modules/$MODULE_ID/lessons \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Video Lesson", "content": "Video content here", "lesson_type": "video", "access_type": "free", "sort_order": 1, "is_preview": true}')
LESSON_ID=$(echo $LESSON_RES | jq -r '.data.id')
echo "Lesson created: $LESSON_ID"

if [ "$LESSON_ID" == "null" ]; then
  echo "Lesson creation failed: $LESSON_RES"
  exit 1
fi

# 3. Generate Video
echo "Generating test video..."
ffmpeg -y -f lavfi -i testsrc=duration=2:size=640x360:rate=30 -c:v libx264 -t 2 test_video.mp4 > /dev/null 2>&1

# 4. Upload Video
echo "Uploading video..."
UPLOAD_RES=$(curl -s -X POST $BASE_URL/videos/lessons/$LESSON_ID/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_video.mp4")

# Check if success
SUCCESS=$(echo $UPLOAD_RES | jq -r '.success')
if [ "$SUCCESS" != "true" ]; then
  echo "Upload failed: $UPLOAD_RES"
  exit 1
fi
VIDEO_ID=$(echo $UPLOAD_RES | jq -r '.data.id')
echo "Upload successful. Video ID: $VIDEO_ID"

# 5. Poll Status
echo "Waiting for processing..."
for i in {1..20}; do
  STATUS_RES=$(curl -s -X GET $BASE_URL/videos/lessons/$LESSON_ID/status \
    -H "Authorization: Bearer $TOKEN")
  STATUS=$(echo $STATUS_RES | jq -r '.data.status')
  echo "Status: $STATUS"
  
  if [ "$STATUS" == "completed" ]; then
    break
  fi
  if [ "$STATUS" == "failed" ]; then
    echo "Processing failed!"
    exit 1
  fi
  sleep 2
done

if [ "$STATUS" != "completed" ]; then
  echo "Processing timed out."
  exit 1
fi

# 6. Get Playback URL
echo "Fetching playback URL..."
PLAYBACK_RES=$(curl -s -X GET "$BASE_URL/videos/lessons/$LESSON_ID/playback?device_id=test-device" \
  -H "Authorization: Bearer $TOKEN")
PLAYBACK_URL=$(echo $PLAYBACK_RES | jq -r '.data.url')

if [ "$PLAYBACK_URL" == "null" ]; then
  echo "Failed to get playback URL: $PLAYBACK_RES"
  exit 1
fi

FULL_URL="http://localhost:8080$PLAYBACK_URL"
echo "Playback URL: $FULL_URL"

# 7. Fetch Manifest
echo "Fetching HLS manifest..."
MANIFEST=$(curl -s "$FULL_URL")

if [[ $MANIFEST == *"#EXTM3U"* ]]; then
  echo "Manifest valid."
  # Check for token in key
  if [[ $MANIFEST == *"?token="* || $MANIFEST == *"&token="* ]]; then
    echo "Token injection verified."
  else
    echo "WARNING: Token not found in manifest segments/keys"
    echo "$MANIFEST"
  fi
else
  echo "Invalid manifest content:"
  echo "$MANIFEST"
  exit 1
fi

echo "Secure Video Pipeline Verification Passed!"
rm test_video.mp4
