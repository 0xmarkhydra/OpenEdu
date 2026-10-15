import { NextRequest, NextResponse } from "next/server";

type RouteContext={params:Promise<{parts:string[]}>};
const uuid=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const quizActions=new Set(["questions","publish","start","submit"]);
async function proxy(request:NextRequest,ctx:RouteContext,method:"GET"|"POST"){
 const {parts}=await ctx.params;
 const allowed=
  parts.length===3&&parts[0]==="classes"&&uuid.test(parts[1])&&parts[2]==="quizzes"||
  parts.length===2&&parts[0]==="quizzes"&&uuid.test(parts[1])&&method==="GET"||
  parts.length===3&&parts[0]==="quizzes"&&uuid.test(parts[1])&&
   (method==="POST"&&quizActions.has(parts[2]) || method==="GET"&&parts[2]==="results");
 if(!allowed)return NextResponse.json({error:"Unknown quiz route"},{status:404});
 if(method!=="GET"&&request.headers.get("origin")!==request.nextUrl.origin)
  return NextResponse.json({error:"Origin verification required"},{status:403});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Authentication required"},{status:401});
 let body:undefined|string;
 if(method==="POST"){
  body=await request.text();
  if(body.length>20000)return NextResponse.json({error:"Payload too large"},{status:413});
 }
 try{
  const upstream=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/"+parts.join("/"),{
   method,cache:"no-store",signal:AbortSignal.timeout(12000),
   headers:{Authorization:"Bearer "+token,"Content-Type":"application/json"},
   body
  });
  return NextResponse.json(await upstream.json(),{
   status:upstream.status,headers:{"Cache-Control":"no-store"}
  });
 }catch{return NextResponse.json({error:"Exam service temporarily unavailable"},{status:503});}
}
export function GET(r:NextRequest,c:RouteContext){return proxy(r,c,"GET");}
export function POST(r:NextRequest,c:RouteContext){return proxy(r,c,"POST");}
