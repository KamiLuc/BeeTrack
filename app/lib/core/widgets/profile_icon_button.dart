import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../api/api_client.dart';
import '../locale/locale_controller.dart';
import '../storage/token_storage.dart';
import '../../features/auth/bloc/auth_bloc.dart';
import '../../l10n/app_localizations.dart';

/// An AppBar action that opens the profile / settings dialog.
/// Drop into any [AppBar.actions] list — it reads all required
/// objects from the widget tree via [context.read].
class ProfileIconButton extends StatelessWidget {
  const ProfileIconButton({super.key});

  @override
  Widget build(BuildContext context) {
    return IconButton(
      icon: const Icon(Icons.account_circle_outlined),
      onPressed: () => showDialog<void>(
        context: context,
        builder: (_) => _ProfileDialog(
          storage: context.read<TokenStorage>(),
          apiClient: context.read<ApiClient>(),
          localeController: context.read<LocaleController>(),
          authBloc: context.read<AuthBloc>(),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------

class _ProfileDialog extends StatefulWidget {
  final TokenStorage storage;
  final ApiClient apiClient;
  final LocaleController localeController;
  final AuthBloc authBloc;

  const _ProfileDialog({
    required this.storage,
    required this.apiClient,
    required this.localeController,
    required this.authBloc,
  });

  @override
  State<_ProfileDialog> createState() => _ProfileDialogState();
}

class _ProfileDialogState extends State<_ProfileDialog> {
  bool _editingName = false;
  late TextEditingController _nameController;
  bool _savingName = false;
  String? _nameError;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.storage.name ?? '');
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _saveName(AppLocalizations l10n) async {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      setState(() => _nameError = l10n.generalRequired);
      return;
    }
    setState(() {
      _nameError = null;
      _savingName = true;
    });
    try {
      await widget.apiClient.dio.patch(
        '/api/v1/users/me/name',
        data: {'name': name},
      );
      await widget.storage.saveName(name);
      if (mounted) {
        setState(() {
          _editingName = false;
          _savingName = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.profileNameUpdated)),
        );
      }
    } on DioException {
      if (mounted) {
        setState(() => _savingName = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final email = widget.storage.email;
    final name = widget.storage.name;
    final currentLocale = widget.localeController.value;

    final isWide = MediaQuery.sizeOf(context).width >= 600;
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380),
        child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.account_circle, size: 72, color: colorScheme.primary),
            const SizedBox(height: 10),
            if (name != null && name.isNotEmpty)
              Text(
                name,
                style: textTheme.titleMedium
                    ?.copyWith(fontWeight: FontWeight.w700),
              ),
            if (email != null && email.isNotEmpty) ...[
              const SizedBox(height: 2),
              Text(
                email,
                style: textTheme.bodySmall
                    ?.copyWith(color: colorScheme.onSurfaceVariant),
              ),
            ],
            const SizedBox(height: 20),
            const Divider(height: 1),
            const SizedBox(height: 12),

            // Language
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(Icons.language,
                        size: 20, color: colorScheme.onSurfaceVariant),
                    const SizedBox(width: 12),
                    Text(l10n.profileLanguage, style: textTheme.bodyMedium),
                  ],
                ),
                const SizedBox(height: 10),
                SizedBox(
                  width: double.infinity,
                  child: SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(
                          value: 'en',
                          label: Text('English',
                              style: TextStyle(fontSize: 12))),
                      ButtonSegment(
                          value: 'pl',
                          label: Text('Polish',
                              style: TextStyle(fontSize: 12))),
                    ],
                    selected: {currentLocale.languageCode},
                    onSelectionChanged: (s) =>
                        widget.localeController.setLocale(Locale(s.first)),
                    style: const ButtonStyle(
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      visualDensity: VisualDensity.compact,
                    ),
                  ),
                ),
              ],
            ),

            const SizedBox(height: 12),
            const Divider(height: 1),
            const SizedBox(height: 4),

            // Display name
            if (_editingName)
              Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const SizedBox(height: 8),
                  TextField(
                    controller: _nameController,
                    decoration: InputDecoration(
                      labelText: l10n.profileDisplayName,
                      border: const OutlineInputBorder(),
                      isDense: true,
                      errorText: _nameError,
                    ),
                    autofocus: true,
                    textInputAction: TextInputAction.done,
                    onChanged: (_) {
                      if (_nameError != null) setState(() => _nameError = null);
                    },
                    onSubmitted: (_) => _saveName(l10n),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton(
                        onPressed: _savingName
                            ? null
                            : () => setState(() => _editingName = false),
                        child: Text(l10n.generalCancel),
                      ),
                      const SizedBox(width: 8),
                      FilledButton(
                        onPressed:
                            _savingName ? null : () => _saveName(l10n),
                        child: _savingName
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                    strokeWidth: 2, color: Colors.white),
                              )
                            : Text(l10n.generalSave),
                      ),
                    ],
                  ),
                ],
              )
            else
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Icon(Icons.badge_outlined,
                    size: 20, color: colorScheme.onSurfaceVariant),
                title:
                    Text(l10n.profileDisplayName, style: textTheme.bodyMedium),
                subtitle: (name != null && name.isNotEmpty)
                    ? Text(name,
                        style: textTheme.bodySmall
                            ?.copyWith(color: colorScheme.onSurfaceVariant))
                    : null,
                trailing: Icon(Icons.edit_outlined,
                    size: 18, color: colorScheme.onSurfaceVariant),
                onTap: () => setState(() {
                  _nameController.text = widget.storage.name ?? '';
                  _editingName = true;
                }),
              ),

            const SizedBox(height: 4),
            const Divider(height: 1),
            const SizedBox(height: 4),

            // Logout
            InkWell(
              onTap: () {
                Navigator.of(context).pop();
                widget.authBloc.add(LogoutRequested());
              },
              focusColor: Colors.transparent,
              hoverColor: colorScheme.error.withAlpha(12),
              splashColor: colorScheme.error.withAlpha(24),
              highlightColor: colorScheme.error.withAlpha(24),
              borderRadius: BorderRadius.circular(4),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child: Row(
                  children: [
                    Icon(Icons.logout, size: 20, color: colorScheme.error),
                    const SizedBox(width: 16),
                    Text(
                      l10n.authLogout,
                      style: textTheme.bodyMedium
                          ?.copyWith(color: colorScheme.error),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
        ),
      ),
    );
  }
}
