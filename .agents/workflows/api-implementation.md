---
description: How to implement new API endpoints in TutorFlow-LMS
---

Follow these steps to implement new API endpoints in the frontend:

### 1. Identify the Backend Endpoint

- Locate the endpoint in `tutorflow-server/internal/handler/`.
- Verify the request/response structure and required authentication.

### 2. Choose the Right Utility

- **Client-side / Client Components**: Use `api` from `@/lib/api`.
  - It handles authentication automatically via cookies.
- **Server Components / SSR**: Use `authServerFetch` or `serverFetch` from `@/lib/api`.
  - `authServerFetch` includes the access token from server-side cookies.

### 3. Implement the Typed Client

- Add new domain logic to `src/lib/api/` (e.g., `user.ts`) or create a new extension.
- Export your API object from the domain file.
- **IMPORTANT**: Re-export your new domain API in `src/lib/api/index.ts`.

Example for a new domain `user.ts`:

```typescript
import { api } from "./client";

export interface UserProfile {
  id: string;
  name: string;
}

export const userApi = {
  getProfile: async () => {
    return api.get<UserProfile>("/auth/me");
  },
};
```

Then in `src/lib/api/index.ts`:

```typescript
export * from "./user";
```

### 4. Usage in Components

All APIs are now easily accessible from a single point:

```typescript
import { authApi, userApi, authServerFetch } from "@/lib/api";
```

- **Server Component example**:

```typescript
import { authServerFetch } from "@/lib/server-api";

export default async function Page() {
  const data = await authServerFetch<MyData>("/my-endpoint");
  // ... render
}
```

- **Client Component example**:

```typescript
"use client";
import { userApi } from "@/lib/user-api";

export function Profile() {
  const handleUpdate = async () => {
    const { data } = await userApi.updateProfile({ name: "New Name" });
    // ... update UI
  };
}
```

### 5. Verification

- Verify the network calls in the browser DevTools.
- Ensure 401 Unauthorized responses correctly redirect to `/login`.
