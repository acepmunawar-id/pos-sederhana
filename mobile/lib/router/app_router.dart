import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/layout/breakpoints.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';
import '../features/account/screens/account_screen.dart';
import '../features/account/screens/close_shift_screen.dart';
import '../features/auth/providers/auth_controller.dart';
import '../features/auth/screens/login_screen.dart';
import '../features/cashier/screens/cart_screen.dart';
import '../features/cashier/screens/cashier_screen.dart';
import '../features/cashier/screens/payment_screen.dart';
import '../features/cashier/screens/scanner_screen.dart';
import '../features/cashier/screens/success_screen.dart';
import '../features/stock/screens/stock_screen.dart';
import '../features/transactions/models/transaction.dart';
import '../features/transactions/screens/transaction_detail_screen.dart';
import '../features/transactions/screens/transactions_screen.dart';

final _rootKey = GlobalKey<NavigatorState>();

/// Re-evaluates redirects whenever the auth state changes.
class _AuthListenable extends ChangeNotifier {
  _AuthListenable(Ref ref) {
    ref.listen(authControllerProvider, (_, __) => notifyListeners());
  }
}

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    navigatorKey: _rootKey,
    initialLocation: '/kasir',
    refreshListenable: _AuthListenable(ref),
    redirect: (context, state) {
      final auth = ref.read(authControllerProvider);
      if (auth.isLoading) return state.matchedLocation == '/splash' ? null : '/splash';
      final loggedIn = auth.valueOrNull != null;
      final atLogin = state.matchedLocation == '/login';
      if (!loggedIn) return atLogin ? null : '/login';
      if (atLogin || state.matchedLocation == '/splash') return '/kasir';
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, __) => const _Splash()),
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => _TabShell(shell: shell),
        branches: [
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/kasir',
              builder: (_, __) => const CashierScreen(),
              routes: [
                GoRoute(path: 'keranjang', parentNavigatorKey: _rootKey, builder: (_, __) => const CartScreen()),
                GoRoute(path: 'bayar', parentNavigatorKey: _rootKey, builder: (_, __) => const PaymentScreen()),
                GoRoute(path: 'scan', parentNavigatorKey: _rootKey, builder: (_, __) => const ScannerScreen()),
                GoRoute(
                  path: 'selesai',
                  parentNavigatorKey: _rootKey,
                  builder: (_, state) => SuccessScreen(transaction: state.extra! as Transaction),
                ),
              ],
            ),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/transaksi',
              builder: (_, __) => const TransactionsScreen(),
              routes: [
                GoRoute(
                  path: ':id',
                  parentNavigatorKey: _rootKey,
                  builder: (_, state) => TransactionDetailScreen(id: int.parse(state.pathParameters['id']!)),
                ),
              ],
            ),
          ]),
          StatefulShellBranch(routes: [GoRoute(path: '/stok', builder: (_, __) => const StockScreen())]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/akun',
              builder: (_, __) => const AccountScreen(),
              routes: [
                GoRoute(path: 'tutup-shift', parentNavigatorKey: _rootKey, builder: (_, __) => const CloseShiftScreen()),
              ],
            ),
          ]),
        ],
      ),
    ],
  );
});

class _Splash extends StatelessWidget {
  const _Splash();

  @override
  Widget build(BuildContext context) =>
      Scaffold(backgroundColor: context.colors.background, body: const Center(child: CupertinoActivityIndicator(radius: 14)));
}

class _Tab {
  const _Tab(this.label, this.icon, this.activeIcon);
  final String label;
  final IconData icon;
  final IconData activeIcon;
}

const _tabs = [
  _Tab('Kasir', CupertinoIcons.cart, CupertinoIcons.cart_fill),
  _Tab('Transaksi', CupertinoIcons.doc_text, CupertinoIcons.doc_text_fill),
  _Tab('Stok', CupertinoIcons.cube_box, CupertinoIcons.cube_box_fill),
  _Tab('Akun', CupertinoIcons.person_crop_circle, CupertinoIcons.person_crop_circle_fill),
];

/// Navigation: bottom tab bar on phones, iPad-style sidebar on tablets/desktop (icon + label always).
class _TabShell extends StatelessWidget {
  const _TabShell({required this.shell});
  final StatefulNavigationShell shell;

  void _go(int i) => shell.goBranch(i, initialLocation: i == shell.currentIndex);

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    if (context.isWide) {
      return Scaffold(
        backgroundColor: c.background,
        body: Row(children: [
          _Sidebar(current: shell.currentIndex, onTap: _go),
          Container(width: 0.5, color: c.separator),
          Expanded(child: shell),
        ]),
      );
    }
    return Scaffold(
      body: shell,
      bottomNavigationBar: CupertinoTabBar(
        currentIndex: shell.currentIndex,
        onTap: _go,
        activeColor: c.primary,
        inactiveColor: c.gray,
        backgroundColor: c.card.withValues(alpha: 0.94),
        height: 56,
        iconSize: 26,
        border: Border(top: BorderSide(color: c.separator.withValues(alpha: 0.6), width: 0.5)),
        items: [
          for (final t in _tabs) BottomNavigationBarItem(icon: Icon(t.icon), activeIcon: Icon(t.activeIcon), label: t.label),
        ],
      ),
    );
  }
}

class _Sidebar extends StatelessWidget {
  const _Sidebar({required this.current, required this.onTap});
  final int current;
  final ValueChanged<int> onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Container(
      width: context.screenSize == ScreenSize.desktop ? 260 : 200,
      color: c.card,
      child: SafeArea(
        right: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.sm, AppSpacing.xl, AppSpacing.sm, AppSpacing.lg),
              child: Row(children: [
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(color: c.primary, borderRadius: BorderRadius.circular(9)),
                  child: Icon(CupertinoIcons.cart_fill, color: c.onAccent, size: 20),
                ),
                const SizedBox(width: AppSpacing.md),
                Text('Kasir POS', style: AppText.headline.copyWith(color: c.text)),
              ]),
            ),
            for (final (i, t) in _tabs.indexed)
              Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: Material(
                  color: i == current ? c.primary : Colors.transparent,
                  borderRadius: BorderRadius.circular(AppRadius.sm),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(AppRadius.sm),
                    onTap: () => onTap(i),
                    child: SizedBox(
                      height: 48,
                      child: Row(children: [
                        const SizedBox(width: AppSpacing.md),
                        Icon(i == current ? t.activeIcon : t.icon, color: i == current ? c.onAccent : c.primary, size: 22),
                        const SizedBox(width: AppSpacing.md),
                        Text(t.label, style: AppText.body.copyWith(color: i == current ? c.onAccent : c.text, fontWeight: FontWeight.w500)),
                      ]),
                    ),
                  ),
                ),
              ),
          ]),
        ),
      ),
    );
  }
}
