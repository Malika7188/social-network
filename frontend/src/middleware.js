import { NextResponse } from "next/server";

// Define protected routes
const protectedRoutes = [
  "/home",
  "/posts",
  "/profile",
  "/messages",
  "/notification",
  "/groups",
  "/events",
  "/Friends",
  "/settings",
];

// Add these to your unprotected routes
const publicRoutes = ["/", "/register", "/forgot-password"];

export function middleware(request) {
  // Get the pathname from the URL
  const { pathname } = request.nextUrl;

  // Check if the path is a public route
  const isPublicRoute = publicRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // If it's a public route, allow access
  if (isPublicRoute) {
    return NextResponse.next();
  }

  // Check if the path is a protected route
  const isProtectedRoute = protectedRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // If it's not a protected route, allow access
  if (!isProtectedRoute) {
    return NextResponse.next();
  }

  // Check for token in cookies
  const token = request.cookies.get("token")?.value;

  // If no token is found and this is a protected route, redirect to login
  if (!token && isProtectedRoute) {

    // Create a response object
    const response = NextResponse.redirect(new URL("/", request.url));

    // Add a custom header to indicate this was a middleware redirect
    // This can be used by client-side code to handle the redirect properly
    response.headers.set("x-middleware-redirect", "true");

    return response;
  }

  // Allow access to protected route with token
  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!api|_next/static|_next/image|favicon.ico).*)",
  ],
};
