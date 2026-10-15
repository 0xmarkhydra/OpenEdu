import { NextRequest, NextResponse } from "next/server";
type Context={params:Promise<{parts:string[]}>};
const uuid=/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i;
async function proxy(request:NextRequest,ctx:Context,method:"GET"|"POST"){
 const {parts}=await ctx.params;
 const valid=(parts.length===1&&parts[0]==="children"&&method==="GET")||
  (parts.length===3&&parts[0]==="children"&&uuid.test(parts[1])&&["grades","attendance","leave"].includes(parts[2])&&(method==="GET"||parts[2]==="leave"));
 if(!valid)return NextResponse.json({error:"Not found"},{status:404});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Authentication required"},{status:401});
 const origin=request.headers.get("origin");
 if(method==="POST"&&(!origin||origin!==request.nextUrl.origin))return NextResponse.json({error:"Invalid origin"},{status:403});
 try{
  const body=method==="POST"?await request.text():undefined;
  if(body&&body.length>12000)return NextResponse.json({error:"Body too large"},{status:413});
  const endpoint=(process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/me/"+parts.join("/");
  const response=await fetch(endpoint,{method,headers:{authorization:"Bearer "+token,"content-type":"application/json"},body,cache:"no-store",signal:AbortSignal.timeout(12000)});
  return NextResponse.json(await response.json(),{status:response.status,headers:{"Cache-Control":"no-store"}});
 }catch{return NextResponse.json({error:"Service unavailable"},{status:503});}
}
export function GET(request:NextRequest,ctx:Context){return proxy(request,ctx,"GET")}
export function POST(request:NextRequest,ctx:Context){return proxy(request,ctx,"POST")}
