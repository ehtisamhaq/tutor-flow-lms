import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const accessToken = request.cookies.get("accessToken")?.value;

  // 1. Route Protection
  const protectedRoutes = [
    "/dashboard",
    "/admin",
    "/tutor",
    "/student",
    "/cart",
    "/learn",
  ];
  const authRoutes = ["/login", "/register"];

  const isProtectedRoute = protectedRoutes.some((route) =>
    pathname.startsWith(route),
  );
  const isAuthRoute = authRoutes.some((route) => pathname.startsWith(route));

  if (isProtectedRoute && !accessToken) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // if (isAuthRoute && accessToken) {
  //   return NextResponse.redirect(new URL("/dashboard", request.url));
  // }

  // 2. Header Injection for API routes (Proxy runs before rewrites)
  if (pathname.startsWith("/api")) {
    const requestHeaders = new Headers(request.headers);
    if (accessToken) {
      requestHeaders.set("Authorization", `Bearer ${accessToken}`);
      // Don't forward cookie header to backend if it's just meant for Next.js
      requestHeaders.delete("cookie");
    }

    // Forward the session_id if expecting it for some reason
    const sessionId = request.cookies.get("session_id")?.value;
    if (sessionId) {
      requestHeaders.set("X-Session-ID", sessionId);
    }

    // In Docker, this hits the 'server' service. Locally, localhost:8080.
    const BACKEND_URL =
      process.env.BACKEND_API_URL || "http://localhost:8080/api/v1";

    const path = pathname.replace(/^\/api/, "");
    const searchParams = request.nextUrl.search;
    const targetUrl = new URL(`${BACKEND_URL}${path}${searchParams}`);

    return NextResponse.rewrite(targetUrl, {
      request: {
        headers: requestHeaders,
      },
    });
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    // Match all paths except static files
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
};
