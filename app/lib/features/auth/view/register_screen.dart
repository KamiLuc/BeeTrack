import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/theme/app_layout.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';
import '../bloc/auth_bloc.dart';

class RegisterScreen extends StatelessWidget {
  const RegisterScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const _RegisterView();
  }
}

class _RegisterView extends StatefulWidget {
  const _RegisterView();

  @override
  State<_RegisterView> createState() => _RegisterViewState();
}

class _RegisterViewState extends State<_RegisterView> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  String? _errorMessage;
  String? _verifiedEmail;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _submit(BuildContext context) {
    if (_formKey.currentState?.validate() ?? false) {
      final email = _emailController.text.trim();
      final lang = Localizations.localeOf(context).languageCode;
      context.read<AuthBloc>().add(
        RegisterSubmitted(
          email: email,
          lang: lang,
          name: email,
          password: _passwordController.text,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthLoading) setState(() => _errorMessage = null);
        if (state is AuthFailure) {
          setState(() => _errorMessage = _mapError(l10n, state.code));
        }
        if (state is AuthVerificationRequired) {
          setState(() => _verifiedEmail = state.email);
        }
      },
      child: Scaffold(
        body: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: AppLayout.formConstraints(context),
                child: _verifiedEmail != null
                    ? _CheckEmailView(email: _verifiedEmail!)
                    : _FormView(
                        formKey: _formKey,
                        emailController: _emailController,
                        passwordController: _passwordController,
                        errorMessage: _errorMessage,
                        onSubmit: _submit,
                      ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  String _mapError(AppLocalizations l10n, String code) {
    return switch (code) {
      'EMAIL_TAKEN' => l10n.authEmailTaken,
      _ => l10n.generalError,
    };
  }
}

class _CheckEmailView extends StatelessWidget {
  final String email;

  const _CheckEmailView({required this.email});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 48),
        const Icon(Icons.mark_email_read_outlined, size: 64),
        const SizedBox(height: 24),
        Text(
          l10n.authCheckEmail,
          style: Theme.of(context).textTheme.headlineMedium,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 12),
        Text(
          l10n.authCheckEmailMessage(email),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 32),
        Center(
          child: TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(l10n.authBackToLogin),
          ),
        ),
      ],
    );
  }
}

class _FormView extends StatelessWidget {
  final TextEditingController emailController;
  final String? errorMessage;
  final GlobalKey<FormState> formKey;
  final void Function(BuildContext) onSubmit;
  final TextEditingController passwordController;

  const _FormView({
    required this.emailController,
    required this.errorMessage,
    required this.formKey,
    required this.onSubmit,
    required this.passwordController,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Form(
      key: formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 48),
          Center(
            child: Text(
              l10n.authRegister,
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ),
          const SizedBox(height: 24),
          TextFormField(
            controller: emailController,
            decoration: InputDecoration(
              labelText: l10n.authEmail,
              counterText: SizeTier.medium.counterText,
            ),
            keyboardType: TextInputType.emailAddress,
            textInputAction: TextInputAction.next,
            maxLength: SizeTier.medium.maxLength,
            validator: (v) {
              if (!_isValidEmail(v)) return l10n.authInvalidEmail;
              return validateSizeTier(v, SizeTier.medium, l10n.authEmail, l10n);
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: passwordController,
            decoration: InputDecoration(
              labelText: l10n.authPassword,
              counterText: '',
            ),
            obscureText: true,
            textInputAction: TextInputAction.done,
            maxLength: maxPasswordLength,
            onFieldSubmitted: (_) => onSubmit(context),
            validator: (v) {
              if (v == null || v.length < 8) return l10n.authWeakPassword;
              return validatePasswordLength(v, l10n);
            },
          ),
          if (errorMessage != null) ...[
            const SizedBox(height: 12),
            Text(
              errorMessage!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
              textAlign: TextAlign.center,
            ),
          ],
          const SizedBox(height: 24),
          BlocBuilder<AuthBloc, AuthState>(
            builder: (context, state) {
              return Center(
                child: SizedBox(
                  width: 200,
                  child: ElevatedButton(
                    onPressed:
                        state is AuthLoading ? null : () => onSubmit(context),
                    child: state is AuthLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : Text(l10n.authRegister),
                  ),
                ),
              );
            },
          ),
          const SizedBox(height: 16),
          Center(
            child: TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: Text(l10n.authHaveAccount),
            ),
          ),
        ],
      ),
    );
  }

  bool _isValidEmail(String? v) {
    if (v == null) return false;
    return RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
    ).hasMatch(v.trim());
  }
}
