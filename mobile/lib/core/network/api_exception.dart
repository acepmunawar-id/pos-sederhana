import 'package:dio/dio.dart';

class FieldError {
  const FieldError(this.field, this.message);
  final String field;
  final String message;
}

/// Error with a user-friendly (Indonesian) message, built from the backend envelope.
class ApiException implements Exception {
  const ApiException(this.message, {this.status = 0, this.errors = const []});

  final String message;
  final int status;
  final List<FieldError> errors;

  bool get isNetwork => status == 0;

  /// Message plus the first field detail, suitable for a toast.
  String get detail => errors.isEmpty ? message : '$message: ${errors.first.message}';

  factory ApiException.fromDio(DioException e) {
    final data = e.response?.data;
    if (data is Map && data['message'] is String) {
      final errs = (data['errors'] as List? ?? [])
          .whereType<Map>()
          .map((m) => FieldError('${m['field']}', '${m['message']}'))
          .toList();
      return ApiException(data['message'] as String, status: e.response?.statusCode ?? 0, errors: errs);
    }
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.sendTimeout:
        return const ApiException('Server terlalu lama merespons. Coba lagi.');
      case DioExceptionType.connectionError:
        return const ApiException('Tidak dapat terhubung ke server. Periksa koneksi internet.');
      default:
        return ApiException('Terjadi kesalahan (${e.response?.statusCode ?? '-'})', status: e.response?.statusCode ?? 0);
    }
  }

  @override
  String toString() => message;
}

/// Turns any thrown error into a message for the UI.
String errorMessage(Object error) => error is ApiException ? error.detail : 'Terjadi kesalahan. Coba lagi.';
