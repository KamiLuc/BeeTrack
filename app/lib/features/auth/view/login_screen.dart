import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../bloc/auth_bloc.dart';
import 'forgot_password_screen.dart';
import 'register_screen.dart';

class LoginScreen extends StatelessWidget {
  /// When provided, the screen shows an app bar with a hamburger that opens
  /// this drawer (used for public navigation). Null when login is the root.
  final Widget? drawer;

  const LoginScreen({super.key, this.drawer});

  @override
  Widget build(BuildContext context) {
    return _LoginView(drawer: drawer);
  }
}

class _LoginView extends StatefulWidget {
  final Widget? drawer;

  const _LoginView({this.drawer});

  @override
  State<_LoginView> createState() => _LoginViewState();
}

class _LoginViewState extends State<_LoginView> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController(
    text: kDebugMode ? 'kamil@op.pl' : null,
  );
  final _passwordController = TextEditingController(
    text: kDebugMode ? 'lion12345' : null,
  );
  String? _errorCode;
  bool _resendLoading = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _submit(BuildContext context) {
    if (_formKey.currentState?.validate() ?? false) {
      context.read<AuthBloc>().add(
        LoginSubmitted(
          email: _emailController.text.trim(),
          password: _passwordController.text,
        ),
      );
    }
  }

  Future<void> _resendVerification(BuildContext context) async {
    setState(() => _resendLoading = true);
    final api = context.read<ApiClient>();
    final lang = Localizations.localeOf(context).languageCode;
    final messenger = ScaffoldMessenger.of(context);
    final checkEmailMsg = AppLocalizations.of(context)!.authCheckEmail;
    try {
      await api.dio.post(
        '/api/v1/auth/resend-verification',
        data: {'email': _emailController.text.trim(), 'lang': lang},
      );
      if (mounted) {
        messenger.showSnackBar(SnackBar(content: Text(checkEmailMsg)));
      }
    } on DioException {
      // best-effort
    } finally {
      if (mounted) setState(() => _resendLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthLoading) setState(() => _errorCode = null);
        if (state is AuthFailure) setState(() => _errorCode = state.code);
      },
      child: Scaffold(
        appBar: widget.drawer != null ? AppBar() : null,
        drawer: widget.drawer,
        body: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: AppLayout.formConstraints(context),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      const SizedBox(height: 48),
                      Center(
                        child: Text(
                          l10n.authLogin,
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                      ),
                      const SizedBox(height: 24),
                      TextFormField(
                        controller: _emailController,
                        decoration: InputDecoration(labelText: l10n.authEmail),
                        keyboardType: TextInputType.emailAddress,
                        textInputAction: TextInputAction.next,
                        validator: (v) => _isValidEmail(v)
                            ? null
                            : l10n.authInvalidEmail,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _passwordController,
                        decoration: InputDecoration(
                          labelText: l10n.authPassword,
                        ),
                        obscureText: true,
                        textInputAction: TextInputAction.done,
                        onFieldSubmitted: (_) => _submit(context),
                        validator: (v) => (v == null || v.length < 8)
                            ? l10n.authWeakPassword
                            : null,
                      ),
                      if (_errorCode != null) ...[
                        const SizedBox(height: 12),
                        Text(
                          _mapError(l10n, _errorCode!),
                          style: TextStyle(
                            color: Theme.of(context).colorScheme.error,
                          ),
                          textAlign: TextAlign.center,
                        ),
                        if (_errorCode == 'EMAIL_NOT_VERIFIED') ...[
                          const SizedBox(height: 4),
                          Center(
                            child: _resendLoading
                                ? const SizedBox(
                                    height: 20,
                                    width: 20,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  )
                                : TextButton(
                                    onPressed: () =>
                                        _resendVerification(context),
                                    child: Text(l10n.authResendEmail),
                                  ),
                          ),
                        ],
                      ],
                      const SizedBox(height: 24),
                      BlocBuilder<AuthBloc, AuthState>(
                        builder: (context, state) {
                          return Center(
                            child: SizedBox(
                              width: 200,
                              child: ElevatedButton(
                                onPressed: state is AuthLoading
                                    ? null
                                    : () => _submit(context),
                                child: state is AuthLoading
                                    ? const SizedBox(
                                        height: 20,
                                        width: 20,
                                        child: CircularProgressIndicator(
                                          strokeWidth: 2,
                                          color: Colors.white,
                                        ),
                                      )
                                    : Text(l10n.authLogin),
                              ),
                            ),
                          );
                        },
                      ),
                      const SizedBox(height: 8),
                      Center(
                        child: TextButton(
                          onPressed: () => Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (_) => const ForgotPasswordScreen(),
                            ),
                          ),
                          child: Text(l10n.authForgotPassword),
                        ),
                      ),
                      Center(
                        child: TextButton(
                          onPressed: () => Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (_) => const RegisterScreen(),
                            ),
                          ),
                          child: Text(l10n.authNoAccount),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  bool _isValidEmail(String? v) {
    if (v == null) return false;
    return RegExp(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$').hasMatch(v.trim());
  }

  String _mapError(AppLocalizations l10n, String code) {
    return switch (code) {
      'INVALID_CREDENTIALS' => l10n.authInvalidCredentials,
      'EMAIL_NOT_VERIFIED' => l10n.authEmailNotVerified,
      _ => l10n.generalError,
    };
  }
}
