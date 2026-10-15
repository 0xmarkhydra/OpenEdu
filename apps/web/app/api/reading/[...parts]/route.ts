import {NextRequest,NextResponse} from "next/server";

type Context={params:Promise<{parts:string[]}>};
const uuid=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

async function proxy(request:NextRequest,ctx:Context,method:"GET"|"PUT"){
 const {parts}=await ctx.params;
 const valid=parts.length===3&&uuid.test(parts[1])&&parts[2]==="progress";
 if(!valid||!(method==="GET"&&parts[0]==="books")&&!(method==="PUT"&&parts[0]==="lessons"))
  return NextResponse.json({error:"Invalid route"},{status:404});
 if(method==="PUT"&&request.headers.get("origin")!==request.nextUrl.origin)
  return NextResponse.json({error:"Origin verification required"},{status:403});
 const token=request.cookies.get("openedu_session")?.value;
 if(!token)return NextResponse.json({error:"Sign in first"},{status:401});
 let body:undefined|string;
 if(method==="PUT"){
  body=await request.text();
  if(body.length>1024)return NextResponse.json({error:"Request too large"},{status:413});
 }
 try{
  const res=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/"+parts.join("/"),{
   method,cache:"no-store",signal:AbortSignal.timeout(10000),
   headers:{"Authorization":"Bearer "+token,"Content-Type":"application/json"},body,
  });
  const payload=await res.json();
  return NextResponse.json(payload,{status:res.status,headers:{"Cache-Control":"no-store"}});
 }catch{return NextResponse.json({error:"Reading service unavailable"},{status:503});}
}
export function GET(r:NextRequest,c:Context){return proxy(r,c,"GET");}
export function PUT(r:NextRequest,c:Context){return proxy(r,c,"PUT");}
