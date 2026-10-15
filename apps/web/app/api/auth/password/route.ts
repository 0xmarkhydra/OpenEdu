import { NextRequest,NextResponse } from "next/server";

export async function POST(request:NextRequest){
 if(request.headers.get("origin")!==request.nextUrl.origin)
  return NextResponse.json({error:"Origin verification required"},{status:403});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Sign in first"},{status:401});
 let body:unknown;
 try{body=await request.json();}catch{return NextResponse.json({error:"Invalid request"},{status:400});}
 if(!body||typeof body!=="object")return NextResponse.json({error:"Invalid request"},{status:400});
 const data=body as Record<string,unknown>;
 if(typeof data.current_password!=="string"||typeof data.new_password!=="string"||
  data.current_password.length>1024||data.new_password.length>1024)
  return NextResponse.json({error:"Invalid password data"},{status:400});
 try{
  const upstream=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/auth/password",{
   method:"POST",cache:"no-store",signal:AbortSignal.timeout(12000),
   headers:{"Authorization":"Bearer "+token,"Content-Type":"application/json"},
   body:JSON.stringify({current_password:data.current_password,new_password:data.new_password})
  });
  const result=await upstream.json();
  if(!upstream.ok)return NextResponse.json({error:result.error||"Password change failed"},{status:upstream.status});
  const response=NextResponse.json({status:"changed"});
  response.cookies.delete("openedu_session");
  return response;
 }catch{return NextResponse.json({error:"Account service temporarily unavailable"},{status:503});}
}
