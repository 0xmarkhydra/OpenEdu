import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const origin = request.headers.get("origin");
  if (origin && new URL(origin).host !== request.nextUrl.host) return NextResponse.json({error:"Invalid origin"},{status:403});
  const token = request.cookies.get("openedu_session")?.value;
  if (!token) return NextResponse.json({error:"Authentication required"},{status:401});
  const { id } = await params;
  if (!/^[0-9a-fA-F-]{36}$/.test(id)) return NextResponse.json({error:"Invalid school"},{status:400});
  try {
    const body=await request.text();
    if(body.length>5000) return NextResponse.json({error:"Request too large"},{status:413});
    const res=await fetch((process.env.API_INTERNAL_URL || "http://localhost:8080") + "/v1/admin/schools/" + encodeURIComponent(id) + "/members", {
      method:"POST",cache:"no-store", signal: AbortSignal.timeout(12000),
      headers:{Authorization:"Bearer "+token,"Content-Type":"application/json"},body,
    });
    return NextResponse.json(await res.json(),{status:res.status});
  } catch { return NextResponse.json({error:"Service unavailable"},{status:503}); }
}
