import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';

class ForgotPasswordScreen extends StatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  State<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends State<ForgotPasswordScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  bool _loading = false;
  bool _sent = false;
  String? _errorMessage;

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;

    setState(() {
      _loading = true;
      _errorMessage = null;
    });

    final api = context.read<ApiClient>();
    final lang = Localizations.localeOf(context).languageCode;

    final errorMsg = AppLocalizations.of(context)!.generalError;
    try {
      await api.dio.post(
        '/api/v1/auth/forgot-password',
        data: {'email': _emailController.text.trim(), 'lang': lang},
      );
      if (mounted) setState(() => _sent = true);
    } on DioException {
      if (mounted) setState(() => _errorMessage = errorMsg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(title: Text(l10n.authForgotPasswordTitle)),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: AppLayout.formConstraints(context),
              child: _sent ? _SentView() : _FormView(
                formKey: _formKey,
                emailController: _emailController,
                errorMessage: _errorMessage,
                loading: _loading,
                onSubmit: _submit,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _SentView extends StatelessWidget {
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
          l10n.authForgotPasswordSent,
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
  final bool loading;
  final VoidCallback onSubmit;

  const _FormView({
    required this.emailController,
    required this.errorMessage,
    required this.formKey,
    required this.loading,
    required this.onSubmit,
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
              l10n.authForgotPasswordTitle,
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            l10n.authForgotPasswordSubtitle,
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 24),
          TextFormField(
            controller: emailController,
            decoration: InputDecoration(
              labelText: l10n.authEmail,
              counterText: SizeTier.medium.counterText,
            ),
            keyboardType: TextInputType.emailAddress,
            textInputAction: TextInputAction.done,
            maxLength: SizeTier.medium.maxLength,
            onFieldSubmitted: (_) => onSubmit(),
            validator: (v) {
              if (!_isValidEmail(v)) return l10n.authInvalidEmail;
              return validateSizeTier(v, SizeTier.medium, l10n.authEmail, l10n);
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
          Center(
            child: SizedBox(
              width: 200,
              child: ElevatedButton(
                onPressed: loading ? null : onSubmit,
                child: loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : Text(l10n.authSendResetLink),
              ),
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
