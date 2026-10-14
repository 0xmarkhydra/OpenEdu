import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) { return proxy(request, "GET"); }
export async function POST(request: NextRequest) { return proxy(request, "POST"); }
async function proxy(request: NextRequest, method: "GET" | "POST") {
  const origin = request.headers.get("origin");
  if (method === "POST" && origin && new URL(origin).host !== request.nextUrl.host) return NextResponse.json({error:"Invalid origin"},{status:403});
  const token = request.cookies.get("openedu_session")?.value;
  if (!token) return NextResponse.json({ error: "Authentication required" }, { status: 401 });
  try {
    const body = method === "POST" ? await request.text() : undefined;
    if (body && body.length > 4000) return NextResponse.json({error:"Request too large"},{status:413});
    const result = await fetch((process.env.API_INTERNAL_URL || "http://localhost:8080") + "/v1/admin/schools", {
      method, cache: "no-store", signal: AbortSignal.timeout(8000),
      headers: { Authorization: "Bearer " + token, "Content-Type": "application/json" },
      body,
    });
    return NextResponse.json(await result.json(), { status: result.status });
  } catch { return NextResponse.json({ error: "Service unavailable" }, { status: 503 }); }
}
