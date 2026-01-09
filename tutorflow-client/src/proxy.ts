import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function proxy(request: NextRequest) {
  // Only proxy requests starting with /api
  if (!request.nextUrl.pathname.startsWith("/api")) {
    return NextResponse.next();
  }

  // Get the path without /api prefix
  const path = request.nextUrl.pathname.replace(/^\/api/, "");
  const searchParams = request.nextUrl.search;

  // Target Backend URL (Go Backend)
  const BACKEND_URL =
    process.env.BACKEND_API_URL || "http://localhost:8080/api/v1";

  // Construct the new URL
  const targetUrl = new URL(`${BACKEND_URL}${path}${searchParams}`);

  // Create the request headers
  const requestHeaders = new Headers(request.headers);

  // Inject Authorization header from cookies if it exists and NOT already present
  const token = request.cookies.get("accessToken")?.value;
  if (token && !requestHeaders.has("Authorization")) {
    requestHeaders.set("Authorization", `Bearer ${token}`);
  }

  // Inject Session ID if it exists
  const sessionId = request.cookies.get("sessionId")?.value;
  if (sessionId && !requestHeaders.has("X-Session-ID")) {
    requestHeaders.set("X-Session-ID", sessionId);
  }

  // Use NextResponse.rewrite to proxy the request
  return NextResponse.rewrite(targetUrl, {
    request: {
      headers: requestHeaders,
    },
  });
}

// Config matcher to only run on /api routes
export const config = {
  matcher: "/api/:path*",
};
