import { NextRequest, NextResponse } from "next/server";

type Context={params:Promise<{parts:string[]}>};
const uuid=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const valid:Record<string,Set<string>>={
  schools:new Set(["years","classes","announcements","overview","guardians"]),
  classes:new Set(["students","teachers","attendance","assignments","timetable","leave-requests"]),
  assignments:new Set(["submissions"])
};

async function proxy(request:NextRequest,ctx:Context,method:"GET"|"POST"|"PUT"|"PATCH"){
  const {parts}=await ctx.params;
  if(parts.length!==3||!Object.prototype.hasOwnProperty.call(valid,parts[0])||!uuid.test(parts[1])||!valid[parts[0]].has(parts[2]))
    return NextResponse.json({error:"Invalid route"},{status:404});
  const origin=request.headers.get("origin");
  if(method!=="GET"&&(!origin||origin!==request.nextUrl.origin)){
    return NextResponse.json({error:"Origin verification required"},{status:403});
  }
  const token=request.cookies.get("openedu_session")?.value;
  if(!token)return NextResponse.json({error:"Sign in first"},{status:401});
  let body:undefined|string;
  if(method!=="GET"){
    body=await request.text();
    if(body.length>24000)return NextResponse.json({error:"Payload too large"},{status:413});
  }
  const url=new URL("/v1/"+parts.join("/"),process.env.API_INTERNAL_URL||"http://localhost:8080");
  if(method==="GET"){
    const date=request.nextUrl.searchParams.get("date");
    if(date)url.searchParams.set("date",date.slice(0,10));
  }
  try{
    const res=await fetch(url,{method,cache:"no-store",signal:AbortSignal.timeout(12000),
      headers:{Authorization:"Bearer "+token,"content-type":"application/json"},body});
    const data=await res.json();
    return NextResponse.json(data,{status:res.status,headers:{"Cache-Control":"no-store"}});
  }catch{return NextResponse.json({error:"API temporarily unavailable"},{status:503});}
}
export function GET(req:NextRequest,ctx:Context){return proxy(req,ctx,"GET");}
export function POST(req:NextRequest,ctx:Context){return proxy(req,ctx,"POST");}
export function PUT(req:NextRequest,ctx:Context){return proxy(req,ctx,"PUT");}
export function PATCH(req:NextRequest,ctx:Context){return proxy(req,ctx,"PATCH");}
