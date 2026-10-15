import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../core/api.dart';

/// Personal reading state only. The API rechecks membership and publication
/// on every write; no progress is written to plaintext device storage.
class LessonDialog extends StatefulWidget {
  final OpenEduApi api;
  final String bookId;
  final String lessonId;
  final String title;
  final String content;
  const LessonDialog({super.key,required this.api,required this.bookId,
    required this.lessonId,required this.title,required this.content});
  @override
  State<LessonDialog> createState()=>_LessonDialogState();
}

class _LessonDialogState extends State<LessonDialog>{
  bool bookmarked=false;
  bool completed=false;
  bool busy=false;
  String? error;
  @override
  void initState(){
    super.initState();
    if(widget.api.signedIn)load();
  }
  Future<void> load()async{
    try{
      final response=await widget.api.request('GET','/v1/books/${widget.bookId}/progress');
      final items=response['items'] as List<dynamic>? ?? [];
      for(final item in items){
        if(item is Map&&item['lesson_id']==widget.lessonId){
          if(mounted)setState((){
            bookmarked=item['bookmarked']==true;
            completed=item['completed']==true;
          });
          break;
        }
      }
    }catch(e){if(mounted)setState(()=>error=e.toString());}
  }
  Future<void> save({bool? isBookmarked,bool? isCompleted})async{
    if(busy)return;
    setState((){busy=true;error=null;});
    try{
      final result=await widget.api.request('PUT','/v1/lessons/${widget.lessonId}/progress',body:{
        'bookmarked':isBookmarked??bookmarked,
        'completed':isCompleted??completed,
      });
      if(mounted)setState((){
        bookmarked=result['bookmarked']==true;
        completed=result['completed']==true;
      });
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  @override
  Widget build(BuildContext context)=>AlertDialog(
    title:Text(widget.title),
    content:SizedBox(width:500,child:SingleChildScrollView(child:Column(
      crossAxisAlignment:CrossAxisAlignment.start,mainAxisSize:MainAxisSize.min,
      children:[
        SelectableText(widget.content,style:const TextStyle(height:1.7)),
        const SizedBox(height:20),
        if(!widget.api.signedIn)const Text('Đăng nhập để lưu bài học và tiến độ đọc.',
          style:TextStyle(color:Colors.blueGrey,fontSize:12)),
        if(error!=null)Text(error!,style:const TextStyle(color:Colors.red)),
      ]))),
    actions:[
      if(widget.api.signedIn)
        TextButton.icon(onPressed:busy?null:()=>save(isBookmarked:!bookmarked),
          icon:Icon(bookmarked?LucideIcons.bookmarkCheck:LucideIcons.bookmark,size:18),
          label:Text(bookmarked?'Đã lưu':'Lưu bài')),
      if(widget.api.signedIn)
        FilledButton.icon(onPressed:busy?null:()=>save(isCompleted:!completed),
          icon:Icon(completed?LucideIcons.checkCheck:LucideIcons.circleCheck,size:18),
          label:Text(completed?'Đã học':'Đánh dấu đã học')),
      TextButton(onPressed:()=>Navigator.pop(context),child:const Text('Đóng')),
    ]);
}
