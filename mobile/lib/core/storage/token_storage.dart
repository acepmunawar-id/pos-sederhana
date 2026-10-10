import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class Tokens {
  const Tokens(this.accessToken, this.refreshToken);
  final String accessToken;
  final String refreshToken;
}

/// Persists JWTs in the platform keychain / keystore.
class TokenStorage {
  TokenStorage([FlutterSecureStorage? storage]) : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;
  static const _access = 'access_token';
  static const _refresh = 'refresh_token';

  Tokens? _cache;

  Future<Tokens?> read() async {
    if (_cache != null) return _cache;
    final a = await _storage.read(key: _access);
    final r = await _storage.read(key: _refresh);
    if (a == null || r == null) return null;
    return _cache = Tokens(a, r);
  }

  Future<void> save(Tokens t) async {
    _cache = t;
    await _storage.write(key: _access, value: t.accessToken);
    await _storage.write(key: _refresh, value: t.refreshToken);
  }

  Future<void> clear() async {
    _cache = null;
    await _storage.delete(key: _access);
    await _storage.delete(key: _refresh);
  }
}
