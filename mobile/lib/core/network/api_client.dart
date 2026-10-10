import 'dart:async';

import 'package:dio/dio.dart';

import '../config/app_config.dart';
import '../storage/token_storage.dart';
import 'api_exception.dart';

class PageMeta {
  const PageMeta({required this.page, required this.limit, required this.total});
  final int page;
  final int limit;
  final int total;

  bool get hasMore => page * limit < total;

  factory PageMeta.fromJson(Map<String, dynamic> j) =>
      PageMeta(page: j['page'] as int, limit: j['limit'] as int, total: j['total'] as int);
}

class Paged<T> {
  const Paged(this.items, this.meta);
  final List<T> items;
  final PageMeta meta;
}

typedef FromJson<T> = T Function(Map<String, dynamic> json);

/// Thin wrapper over Dio that unwraps `{success, data, meta}` and refreshes JWTs.
class ApiClient {
  ApiClient(this._tokens, {Dio? dio}) : _dio = dio ?? Dio() {
    _dio.options
      ..baseUrl = AppConfig.apiUrl
      ..connectTimeout = const Duration(seconds: 15)
      ..receiveTimeout = const Duration(seconds: 30);
    _dio.interceptors.add(_AuthInterceptor(this));
  }

  final Dio _dio;
  final TokenStorage _tokens;

  /// Invoked when the refresh token is rejected; the app returns to the login screen.
  void Function()? onSessionExpired;

  Completer<String?>? _refreshing;

  Future<T> get<T>(String path, FromJson<T> parse, {Map<String, dynamic>? query}) async {
    final res = await _send(() => _dio.get<Map<String, dynamic>>(path, queryParameters: _clean(query)));
    return parse(res['data'] as Map<String, dynamic>);
  }

  Future<List<T>> getList<T>(String path, FromJson<T> parse, {Map<String, dynamic>? query}) async {
    final res = await _send(() => _dio.get<Map<String, dynamic>>(path, queryParameters: _clean(query)));
    return (res['data'] as List).cast<Map<String, dynamic>>().map(parse).toList();
  }

  Future<Paged<T>> getPage<T>(String path, FromJson<T> parse, {Map<String, dynamic>? query}) async {
    final res = await _send(() => _dio.get<Map<String, dynamic>>(path, queryParameters: _clean(query)));
    final items = (res['data'] as List).cast<Map<String, dynamic>>().map(parse).toList();
    final meta = res['meta'] is Map<String, dynamic>
        ? PageMeta.fromJson(res['meta'] as Map<String, dynamic>)
        : PageMeta(page: 1, limit: items.length, total: items.length);
    return Paged(items, meta);
  }

  /// POST returning the raw envelope, so callers can read `data` and status-dependent fields.
  Future<Map<String, dynamic>> post(String path, {Object? body, Map<String, String>? headers}) =>
      _send(() => _dio.post<Map<String, dynamic>>(path, data: body, options: Options(headers: headers)));

  Future<Map<String, dynamic>> put(String path, {Object? body}) =>
      _send(() => _dio.put<Map<String, dynamic>>(path, data: body));

  Future<Map<String, dynamic>> _send(Future<Response<Map<String, dynamic>>> Function() call) async {
    try {
      final res = await call();
      return res.data ?? const {};
    } on DioException catch (e) {
      throw e.error is ApiException ? e.error as ApiException : ApiException.fromDio(e);
    }
  }

  Map<String, dynamic>? _clean(Map<String, dynamic>? q) {
    if (q == null) return null;
    return Map.fromEntries(q.entries.where((e) => e.value != null && e.value != ''));
  }

  /// Returns a new access token or null if the session cannot be refreshed.
  /// Concurrent callers share a single refresh request.
  Future<String?> _refreshAccessToken() async {
    if (_refreshing != null) return _refreshing!.future;
    final completer = _refreshing = Completer<String?>();
    try {
      final tokens = await _tokens.read();
      if (tokens == null) {
        completer.complete(null);
      } else {
        final res = await Dio(BaseOptions(baseUrl: AppConfig.apiUrl))
            .post<Map<String, dynamic>>('/auth/refresh', data: {'refresh_token': tokens.refreshToken});
        final data = res.data!['data'] as Map<String, dynamic>;
        final next = Tokens(data['access_token'] as String, data['refresh_token'] as String);
        await _tokens.save(next);
        completer.complete(next.accessToken);
      }
    } catch (_) {
      await _tokens.clear();
      completer.complete(null);
      onSessionExpired?.call();
    } finally {
      _refreshing = null;
    }
    return completer.future;
  }

  Future<void> saveTokens(Tokens t) => _tokens.save(t);
  Future<Tokens?> readTokens() => _tokens.read();
  Future<void> clearTokens() => _tokens.clear();
}

class _AuthInterceptor extends Interceptor {
  _AuthInterceptor(this.client);
  final ApiClient client;

  @override
  Future<void> onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    final tokens = await client._tokens.read();
    if (tokens != null) options.headers['Authorization'] = 'Bearer ${tokens.accessToken}';
    handler.next(options);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    final path = err.requestOptions.path;
    final retried = err.requestOptions.extra['retried'] == true;
    if (err.response?.statusCode == 401 && !retried && !path.contains('/auth/login') && !path.contains('/auth/refresh')) {
      final token = await client._refreshAccessToken();
      if (token != null) {
        final opts = err.requestOptions
          ..headers['Authorization'] = 'Bearer $token'
          ..extra['retried'] = true;
        try {
          return handler.resolve(await client._dio.fetch(opts));
        } on DioException catch (e) {
          return handler.next(e);
        }
      }
    }
    handler.next(err);
  }
}
