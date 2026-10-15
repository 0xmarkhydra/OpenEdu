import 'dart:async';
import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../core/api.dart';

/// Student-facing native exam client. Server remains sole authority on time
/// limits, answer keys and scores; this view is only presentation.
class QuizScreen extends StatefulWidget {
  final OpenEduApi api;
  final String classId;
  final String className;
  const QuizScreen({super.key,required this.api,required this.classId,required this.className});
  @override
  State<QuizScreen> createState()=>_QuizScreenState();
}
class _QuizScreenState extends State<QuizScreen>{
  List<Map<String,dynamic>> quizzes=[];
  Map<String,dynamic>? detail;
  Map<String,int> answers={};
  String? error;
  bool busy=false;
  DateTime now=DateTime.now();
  Timer? clock;
  @override
  void initState(){super.initState();clock=Timer.periodic(const Duration(seconds:1),(_){
    if(mounted)setState(()=>now=DateTime.now());
  });refresh();}
  @override
  void dispose(){clock?.cancel();super.dispose();}
  Future<void> refresh()async{
    setState((){busy=true;error=null;});
    try{
      final res=await widget.api.request('GET','/v1/classes/${widget.classId}/quizzes');
      if(mounted)setState(()=>quizzes=(res['items'] as List<dynamic>? ?? [])
        .map((v)=>Map<String,dynamic>.from(v as Map)).toList());
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  Future<void> open(String id)async{
    setState((){busy=true;error=null;answers={};});
    try{
      final res=await widget.api.request('GET','/v1/quizzes/$id');
      if(mounted)setState(()=>detail=res);
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  Future<void> start()async{
    final id=detail?['id']?.toString();
    if(id==null)return;
    setState((){busy=true;error=null;});
    try{
      await widget.api.request('POST','/v1/quizzes/$id/start',body:{});
      await open(id);
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  Future<void> submit()async{
    final id=detail?['id']?.toString();
    if(id==null)return;
    setState((){busy=true;error=null;});
    try{
      await widget.api.request('POST','/v1/quizzes/$id/submit',body:{'answers':answers});
      await open(id);
      await refresh();
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  bool get staff=>widget.api.roles.any((r)=>['teacher','principal','school_admin'].contains(r));
  Widget body(){
    if(detail==null)return Column(crossAxisAlignment:CrossAxisAlignment.stretch,children:[
      const Text('Bài kiểm tra',style:TextStyle(fontSize:27,fontWeight:FontWeight.w800)),
      Text(widget.className,style:const TextStyle(color:Colors.blueGrey)),
      const SizedBox(height:18),
      if(quizzes.isEmpty)const Text('Chưa có bài kiểm tra được phát hành cho lớp học này.'),
      ...quizzes.map((q)=>Card(
        child:ListTile(
          leading:const Icon(LucideIcons.clipboardList,color:Color(0xFF5676D5)),
          title:Text(q['title']?.toString()??'Đề kiểm tra'),
          subtitle:Text('${q['duration_minutes']} phút · ${q['questions']} câu' +
            (q['grade']==null?'':' · Điểm ${q['grade']}/10')),
          trailing:const Icon(LucideIcons.chevronRight),
          onTap:()=>open(q['id'].toString()),
        ),
      ))
    ]);
    final d=detail!;
    final items=(d['items'] as List<dynamic>? ?? [])
      .map((e)=>Map<String,dynamic>.from(e as Map)).toList();
    final submitted=d['submitted']==true;
    final started=d['started']==true;
    final expiry=DateTime.tryParse(d['expires_at']?.toString()??'');
    final remaining=expiry?.difference(now)??Duration.zero;
    final timeValid=remaining.inSeconds>0;
    return Column(crossAxisAlignment:CrossAxisAlignment.stretch,children:[
      TextButton.icon(onPressed:()=>setState(()=>detail=null),
        icon:const Icon(LucideIcons.arrowLeft,size:17),label:const Text('Danh sách đề')),
      const SizedBox(height:10),
      const Text('Làm bài kiểm tra',style:TextStyle(fontSize:25,fontWeight:FontWeight.w800)),
      const SizedBox(height:15),
      if(submitted)Card(child:Padding(padding:const EdgeInsets.all(20),child:Column(children:[
        const Icon(LucideIcons.circleCheck,size:36,color:Color(0xFF3E8BCA)),
        const SizedBox(height:10),const Text('Đã nộp bài',style:TextStyle(fontWeight:FontWeight.w700)),
        Text('Điểm: ${d['grade']??'Chưa có'}/10',style:const TextStyle(fontSize:24,fontWeight:FontWeight.w800))
      ]))),
      if(!submitted&&!staff&&!started)FilledButton.icon(onPressed:busy?null:start,
        icon:const Icon(LucideIcons.play,size:18),label:const Text('Bắt đầu làm bài')),
      if(!submitted&&!staff&&started) ...[
        Text(timeValid?'Còn ${(remaining.inSeconds/60).ceil()} phút':'Đã hết giờ',
          style:const TextStyle(color:Color(0xFF506EC8),fontWeight:FontWeight.w700)),
        if(timeValid) ...[
          for(final q in items)Card(child:Padding(padding:const EdgeInsets.all(14),
            child:Column(crossAxisAlignment:CrossAxisAlignment.start,children:[
              Text('${q['position']}. ${q['prompt']}',style:const TextStyle(fontWeight:FontWeight.w700)),
              const SizedBox(height:10),
              for(var i=0;i<(q['options'] as List<dynamic>? ?? []).length;i++)
                RadioListTile<int>(
                  dense:true,
                  title:Text('${String.fromCharCode(65+i)}. ${(q['options'] as List<dynamic>)[i]}'),
                  value:i,
                  groupValue:answers[q['id'].toString()],
                  onChanged:(v)=>setState(()=>answers[q['id'].toString()]=v??0),
                ),
            ]))),
          const SizedBox(height:12),
          FilledButton.icon(
            onPressed:busy||items.isEmpty||items.any((q)=>!answers.containsKey(q['id'].toString()))?null:submit,
            icon:const Icon(LucideIcons.send,size:18),label:const Text('Nộp bài')),
        ],
      ],
      if(staff)const Text('Quản lý soạn đề và xem báo cáo chi tiết trên giao diện web OpenEdu.'),
    ]);
  }
  @override
  Widget build(BuildContext context)=>Scaffold(
    appBar:AppBar(title:const Text('OpenEdu · Kiểm tra')),
    body:SafeArea(child:RefreshIndicator(onRefresh:refresh,child:ListView(
      padding:const EdgeInsets.all(18),children:[
       if(busy)const LinearProgressIndicator(),
       if(error!=null)Padding(padding:const EdgeInsets.all(12),
         child:Text(error!,style:const TextStyle(color:Colors.red))),
       body()
      ]
    )))
  );
}
