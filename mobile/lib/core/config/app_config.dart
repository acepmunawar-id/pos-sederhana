/// Build-time configuration, injected with `--dart-define-from-file=.env`.
class AppConfig {
  AppConfig._();

  static const String apiUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://10.0.2.2:8080/api/v1',
  );

  /// Scheme + host + port of the API, used to fix image URLs that point at "localhost".
  static Uri get apiOrigin => Uri.parse(apiUrl).replace(path: '', query: null);
}
