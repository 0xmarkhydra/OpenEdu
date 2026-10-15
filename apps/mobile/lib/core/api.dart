import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

/// Configure with --dart-define=OPENEDU_API_URL=https://your-approved-api-domain.
/// Production rejects plaintext HTTP, and never ships embedded admin credentials.
class OpenEduApi {
  static const endpoint = String.fromEnvironment('OPENEDU_API_URL');
  final http.Client _client;
  final FlutterSecureStorage _storage;
  String? _token;
  String? schoolId;
  String? userId;
  List<String> roles = [];

  OpenEduApi({http.Client? client, FlutterSecureStorage? storage})
      : _client = client ?? http.Client(),
        _storage = storage ?? const FlutterSecureStorage();

  bool get configured => endpoint.startsWith('https://') ||
      (kDebugMode && endpoint.startsWith('http://localhost'));
  bool get signedIn => _token != null && _token!.isNotEmpty;

  Future<void> restoreSession() async {
    _token = await _storage.read(key: 'openedu_access_token');
    if (_token == null) return;
    try {
      await me();
    } catch (_) {
      await signOut();
    }
  }

  Uri _uri(String route, [Map<String, String>? query]) {
    if (!configured) throw const ApiException(
      'API mobile chưa được cấu hình. Cần cấp URL HTTPS riêng cho ứng dụng.');
    return Uri.parse('${endpoint.replaceAll(RegExp(r"/+$"), "")}$route')
        .replace(queryParameters: query);
  }

  Future<Map<String, dynamic>> request(
    String method,
    String route, {
    Map<String, String>? query,
    Map<String, dynamic>? body,
  }) async {
    final uri = _uri(route, query);
    try {
      final response = await _client.send(http.Request(method, uri)
        ..headers.addAll({
          'Content-Type': 'application/json',
          if (signedIn) 'Authorization': 'Bearer $_token',
        })
        ..body = body == null ? '' : jsonEncode(body)).timeout(
          const Duration(seconds: 12));
      final result = await http.Response.fromStream(response);
      final decoded = jsonDecode(result.body);
      if (decoded is! Map<String, dynamic>) {
        throw const ApiException('Phản hồi API không hợp lệ.');
      }
      if (result.statusCode < 200 || result.statusCode >= 300) {
        throw ApiException(decoded['error']?.toString() ??
            'Máy chủ từ chối yêu cầu (${result.statusCode}).');
      }
      return decoded;
    } on ApiException {
      rethrow;
    } catch (_) {
      throw const ApiException('Không thể kết nối tới máy chủ OpenEdu.');
    }
  }

  Future<void> login(String email, String password, {String schoolId = ''}) async {
    final result = await request('POST', '/v1/auth/login', body: {
      'email': email.trim(),
      'password': password,
      'school_id': schoolId.trim(),
    });
    final token = result['access_token'];
    if (token is! String || token.isEmpty) {
      throw const ApiException('Phiên đăng nhập không hợp lệ.');
    }
    await _storage.write(key: 'openedu_access_token', value: token);
    _token = token;
    this.schoolId = result['school_id']?.toString();
    userId = result['user_id']?.toString();
    roles = (result['roles'] as List<dynamic>? ?? []).cast<String>();
  }

  Future<void> me() async {
    final profile = await request('GET', '/v1/auth/me');
    schoolId = profile['school_id']?.toString();
    userId = profile['user_id']?.toString();
    roles = (profile['roles'] as List<dynamic>? ?? []).cast<String>();
  }

  Future<void> signOut() async {
    _token = null;
    schoolId = null;
    userId = null;
    roles = [];
    await _storage.delete(key: 'openedu_access_token');
  }

  Future<List<Map<String, dynamic>>> books({String query = ''}) async {
    final result = await request('GET', '/v1/books', query: {
      if (query.trim().isNotEmpty) 'q': query.trim(),
    });
    return (result['items'] as List<dynamic>? ?? [])
        .map((item) => Map<String, dynamic>.from(item as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> schoolClasses() async {
    if (schoolId == null || schoolId!.isEmpty) return [];
    final result = await request('GET', '/v1/schools/$schoolId/classes');
    return (result['items'] as List<dynamic>? ?? [])
        .map((item) => Map<String, dynamic>.from(item as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> announcements() async {
    if (schoolId == null || schoolId!.isEmpty) return [];
    final result = await request('GET', '/v1/schools/$schoolId/announcements');
    return (result['items'] as List<dynamic>? ?? [])
        .map((item) => Map<String, dynamic>.from(item as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> messages() async {
    final result = await request('GET', '/v1/me/messages');
    return (result['items'] as List<dynamic>? ?? [])
        .map((item) => Map<String, dynamic>.from(item as Map)).toList();
  }

  void dispose() => _client.close();
}

class ApiException implements Exception {
  final String message;
  const ApiException(this.message);
  @override
  String toString() => message;
}
