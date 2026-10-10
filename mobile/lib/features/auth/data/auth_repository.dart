import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/providers.dart';
import '../../../core/storage/token_storage.dart';
import '../models/user.dart';

class AuthRepository {
  AuthRepository(this._api);
  final ApiClient _api;

  Future<AppUser> login(String email, String password) async {
    final res = await _api.post('/auth/login', body: {'email': email.trim(), 'password': password});
    final data = res['data'] as Map<String, dynamic>;
    await _api.saveTokens(Tokens(data['access_token'] as String, data['refresh_token'] as String));
    return AppUser.fromJson(data['user'] as Map<String, dynamic>);
  }

  /// Returns the signed-in user, or null when there is no stored session.
  Future<AppUser?> restore() async {
    if (await _api.readTokens() == null) return null;
    return _api.get('/auth/me', AppUser.fromJson);
  }

  Future<void> logout() async {
    final tokens = await _api.readTokens();
    try {
      if (tokens != null) await _api.post('/auth/logout', body: {'refresh_token': tokens.refreshToken});
    } finally {
      await _api.clearTokens();
    }
  }

  Future<void> changePassword(String oldPassword, String newPassword) async {
    await _api.put('/auth/password', body: {'old_password': oldPassword, 'new_password': newPassword});
    await _api.clearTokens();
  }
}

final authRepositoryProvider = Provider((ref) => AuthRepository(ref.watch(apiClientProvider)));
