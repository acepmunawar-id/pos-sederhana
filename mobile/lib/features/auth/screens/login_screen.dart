import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/haptics.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/inputs.dart';
import '../providers/auth_controller.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  bool _loading = false;
  String? _error;
  String? _emailError;
  String? _passwordError;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    FocusScope.of(context).unfocus();
    setState(() {
      _error = null;
      _emailError = _email.text.trim().isEmpty ? 'Email wajib diisi' : null;
      _passwordError = _password.text.isEmpty ? 'Password wajib diisi' : null;
    });
    if (_emailError != null || _passwordError != null) {
      Haptics.error();
      return;
    }
    setState(() => _loading = true);
    try {
      await ref.read(authControllerProvider.notifier).login(_email.text, _password.text);
      Haptics.success();
    } on ApiException catch (e) {
      Haptics.error();
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      backgroundColor: c.background,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: AutofillGroup(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Center(
                      child: Container(
                        width: 84,
                        height: 84,
                        decoration: BoxDecoration(color: c.primary, borderRadius: BorderRadius.circular(20)),
                        child: Icon(CupertinoIcons.cart_fill, color: c.onAccent, size: 42),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.xl),
                    Text('Kasir POS', textAlign: TextAlign.center, style: AppText.largeTitle.copyWith(color: c.text)),
                    const SizedBox(height: AppSpacing.xs),
                    Text('Masuk untuk mulai berjualan', textAlign: TextAlign.center, style: AppText.body.copyWith(color: c.textSecondary)),
                    const SizedBox(height: AppSpacing.xxl),
                    if (_error != null) ...[
                      Container(
                        padding: const EdgeInsets.all(AppSpacing.md),
                        decoration: BoxDecoration(color: c.danger.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(AppRadius.sm)),
                        child: Row(children: [
                          Icon(CupertinoIcons.exclamationmark_circle_fill, color: c.danger, size: 20),
                          const SizedBox(width: AppSpacing.sm),
                          Expanded(child: Text(_error!, style: AppText.subheadline.copyWith(color: c.danger))),
                        ]),
                      ),
                      const SizedBox(height: AppSpacing.lg),
                    ],
                    AppTextField(
                      label: 'Email',
                      controller: _email,
                      placeholder: 'nama@toko.id',
                      keyboardType: TextInputType.emailAddress,
                      textInputAction: TextInputAction.next,
                      autofillHints: const [AutofillHints.email],
                      error: _emailError,
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    AppTextField(
                      label: 'Password',
                      controller: _password,
                      placeholder: 'Password',
                      obscure: true,
                      textInputAction: TextInputAction.done,
                      autofillHints: const [AutofillHints.password],
                      onSubmitted: (_) => _submit(),
                      error: _passwordError,
                    ),
                    const SizedBox(height: AppSpacing.xl),
                    AppButton(label: 'Masuk', icon: CupertinoIcons.arrow_right_circle_fill, loading: _loading, onPressed: _submit),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
