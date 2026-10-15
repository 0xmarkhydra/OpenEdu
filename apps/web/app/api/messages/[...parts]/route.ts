import { NextRequest, NextResponse } from "next/server";
type Params={params:Promise<{parts:string[]}>};
const uuid=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

async function proxy(request:NextRequest,context:Params,method:"GET"|"POST"|"PATCH"){
 const {parts}=await context.params;
 const route=parts.length===1 && parts[0]==="inbox" ? "" :
  parts.length===1 && parts[0]==="contacts" ? "/contacts" :
  parts.length===2 && uuid.test(parts[0]) && parts[1]==="read" ? "/"+parts[0]+"/read" : null;
 if(route===null || (method==="GET" && route.endsWith("/read")) ||
   (method==="POST" && route!=="") ||
   (method==="PATCH" && !route.endsWith("/read")))
   return NextResponse.json({error:"Invalid route"},{status:404});
 if(method!=="GET"&&request.headers.get("origin")!==request.nextUrl.origin)
   return NextResponse.json({error:"Origin check failed"},{status:403});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Authentication required"},{status:401});
 let body:undefined|string;
 if(method==="POST"){
  body=await request.text();
  if(body.length>3000)return NextResponse.json({error:"Message too large"},{status:413});
 }
 try{
  const res=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/me/messages"+route,{
   method,cache:"no-store",signal:AbortSignal.timeout(10000),
   headers:{"Authorization":"Bearer "+token,"Content-Type":"application/json"},body,
  });
  const data=await res.json();
  return NextResponse.json(data,{status:res.status,headers:{"Cache-Control":"no-store"}});
 }catch{return NextResponse.json({error:"Messaging service unavailable"},{status:503});}
}
export function GET(r:NextRequest,c:Params){return proxy(r,c,"GET");}
export function POST(r:NextRequest,c:Params){return proxy(r,c,"POST");}
export function PATCH(r:NextRequest,c:Params){return proxy(r,c,"PATCH");}
