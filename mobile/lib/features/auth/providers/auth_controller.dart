import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers.dart';
import '../../cart/providers/cart_controller.dart';
import '../data/auth_repository.dart';
import '../models/user.dart';

/// Holds the signed-in user (null = signed out).
class AuthController extends AsyncNotifier<AppUser?> {
  @override
  Future<AppUser?> build() async {
    ref.read(apiClientProvider).onSessionExpired = _expire;
    try {
      return await ref.read(authRepositoryProvider).restore();
    } catch (_) {
      await ref.read(apiClientProvider).clearTokens();
      return null;
    }
  }

  Future<void> login(String email, String password) async {
    final user = await ref.read(authRepositoryProvider).login(email, password);
    state = AsyncData(user);
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    _reset();
  }

  Future<void> changePassword(String oldPassword, String newPassword) async {
    await ref.read(authRepositoryProvider).changePassword(oldPassword, newPassword);
    _reset();
  }

  void _expire() => _reset();

  void _reset() {
    ref.read(cartControllerProvider.notifier).clear();
    state = const AsyncData(null);
  }
}

final authControllerProvider = AsyncNotifierProvider<AuthController, AppUser?>(AuthController.new);

/// Convenience: the current user (throws if not signed in; only use behind the auth guard).
final currentUserProvider = Provider<AppUser>((ref) => ref.watch(authControllerProvider).requireValue!);
