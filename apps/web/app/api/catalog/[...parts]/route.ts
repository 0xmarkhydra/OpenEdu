import { NextRequest,NextResponse } from "next/server";
type Context={params:Promise<{parts:string[]}>};
const uuid=/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i;
const actions=new Set(["submit","approve","publish","archive"]);
async function proxy(request:NextRequest,ctx:Context,method:"GET"|"POST"){
 const {parts}=await ctx.params;
 const valid=
   parts.length===1&&parts[0]==="books"||
   parts.length===3&&parts[0]==="books"&&uuid.test(parts[1])&&((parts[2]==="chapters")||method==="POST"&&actions.has(parts[2]))||
   parts.length===3&&parts[0]==="chapters"&&uuid.test(parts[1])&&parts[2]==="lessons"&&method==="POST";
 if(!valid)return NextResponse.json({error:"Not found"},{status:404});
 const origin=request.headers.get("origin");
 if(method==="POST"&&(!origin||origin!==request.nextUrl.origin))return NextResponse.json({error:"CSRF protection"},{status:403});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Authentication required"},{status:401});
 try{
  const body=method==="POST"?await request.text():undefined;
  if(body&&body.length>65000)return NextResponse.json({error:"Payload too large"},{status:413});
  const res=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/admin/"+parts.join("/"),{
   method,cache:"no-store",signal:AbortSignal.timeout(10000),
   headers:{"authorization":"Bearer "+token,"content-type":"application/json"},body});
  return NextResponse.json(await res.json(),{status:res.status,headers:{"Cache-Control":"no-store"}});
 }catch{return NextResponse.json({error:"Editorial API unavailable"},{status:503});}
}
export function GET(request:NextRequest,ctx:Context){return proxy(request,ctx,"GET")}
export function POST(request:NextRequest,ctx:Context){return proxy(request,ctx,"POST")}
