import { NextRequest, NextResponse } from "next/server";
export async function POST(request: NextRequest) {
  const origin = request.headers.get("origin");
  if (origin && new URL(origin).host !== request.nextUrl.host) return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  const response = NextResponse.redirect(new URL("/login", request.url), { status: 303 });
  response.cookies.delete("openedu_session");
  return response;
}
