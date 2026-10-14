import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const origin = request.headers.get("origin");
  if (origin && new URL(origin).host !== request.nextUrl.host) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }
  let body: unknown;
  try { body = await request.json(); } catch {
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
  }
  if (!body || typeof body !== "object") return NextResponse.json({ error: "Invalid request" }, { status: 400 });
  const data = body as Record<string, unknown>;
  if (typeof data.email !== "string" || typeof data.password !== "string" || data.email.length > 254 || data.password.length > 1024 ||
      (data.school_id != null && typeof data.school_id !== "string")) {
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
  }
  try {
    const upstream = await fetch((process.env.API_INTERNAL_URL || "http://localhost:8080") + "/v1/auth/login", {
      method: "POST", cache: "no-store", signal: AbortSignal.timeout(12000),
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: data.email, password: data.password, school_id: data.school_id || "" }),
    });
    const payload = await upstream.json();
    if (!upstream.ok) return NextResponse.json({ error: payload.error || "Authentication failed" }, { status: upstream.status });
    const { access_token, token_type, expires_in, ...profile } = payload;
    if (typeof access_token !== "string" || !access_token) throw new Error("invalid session");
    const response = NextResponse.json({ ...profile, authenticated: true });
    response.cookies.set("openedu_session", access_token, {
      httpOnly: true, secure: process.env.NODE_ENV === "production", sameSite: "strict",
      path: "/", maxAge: Math.min(Number(expires_in) || 900, 900),
    });
    return response;
  } catch {
    return NextResponse.json({ error: "Login service unavailable" }, { status: 503 });
  }
}
