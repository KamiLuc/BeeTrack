import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../api/api_client.dart';
import '../locale/locale_controller.dart';
import '../storage/token_storage.dart';
import '../validation/size_tiers.dart';
import '../../features/apiary/data/invitation_model.dart';
import '../../features/apiary/data/invitation_repository.dart';
import '../../features/auth/bloc/auth_bloc.dart';
import '../../l10n/app_localizations.dart';

/// An AppBar action that shows a notification badge for pending invitations
/// and opens the profile / settings dialog on tap.
class ProfileIconButton extends StatefulWidget {
  final VoidCallback? onRefresh;

  const ProfileIconButton({super.key, this.onRefresh});

  @override
  State<ProfileIconButton> createState() => _ProfileIconButtonState();
}

class _ProfileIconButtonState extends State<ProfileIconButton> {
  int _pendingCount = 0;

  @override
  void initState() {
    super.initState();
    _fetchCount();
  }

  Future<void> _fetchCount() async {
    try {
      final count = await InvitationRepository(api: context.read<ApiClient>()).countMine();
      if (mounted) setState(() => _pendingCount = count);
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        IconButton(
          icon: const Icon(Icons.account_circle_outlined),
          onPressed: () async {
            await showDialog<void>(
              context: context,
              builder: (_) => _ProfileDialog(
                storage: context.read<TokenStorage>(),
                apiClient: context.read<ApiClient>(),
                localeController: context.read<LocaleController>(),
                authBloc: context.read<AuthBloc>(),
                onAccepted: widget.onRefresh,
              ),
            );
            _fetchCount();
          },
        ),
        if (_pendingCount > 0)
          Positioned(
            right: 6,
            top: 6,
            child: IgnorePointer(
              child: Container(
                width: 10,
                height: 10,
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.error,
                  shape: BoxShape.circle,
                ),
              ),
            ),
          ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------

class _ProfileDialog extends StatefulWidget {
  final TokenStorage storage;
  final ApiClient apiClient;
  final LocaleController localeController;
  final AuthBloc authBloc;
  final VoidCallback? onAccepted;

  const _ProfileDialog({
    required this.storage,
    required this.apiClient,
    required this.localeController,
    required this.authBloc,
    this.onAccepted,
  });

  @override
  State<_ProfileDialog> createState() => _ProfileDialogState();
}

class _ProfileDialogState extends State<_ProfileDialog> {
  bool _editingName = false;
  late TextEditingController _nameController;
  bool _savingName = false;
  String? _nameError;

  List<MyInvitation>? _invitations;
  final Set<int> _actioning = {};

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.storage.name ?? '');
    _loadInvitations();
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _loadInvitations() async {
    try {
      final list = await InvitationRepository(api: widget.apiClient).listMine();
      if (mounted) setState(() => _invitations = list);
    } catch (_) {}
  }

  Future<void> _accept(int id) async {
    setState(() => _actioning.add(id));
    try {
      await InvitationRepository(api: widget.apiClient).accept(id);
      widget.onAccepted?.call();
      await _loadInvitations();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.generalError)),
        );
      }
    } finally {
      if (mounted) setState(() => _actioning.remove(id));
    }
  }

  Future<void> _decline(int id) async {
    setState(() => _actioning.add(id));
    try {
      await InvitationRepository(api: widget.apiClient).decline(id);
      await _loadInvitations();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.generalError)),
        );
      }
    } finally {
      if (mounted) setState(() => _actioning.remove(id));
    }
  }

  Future<void> _saveName(AppLocalizations l10n) async {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      setState(() => _nameError = l10n.generalRequired);
      return;
    }
    final tooLong = validateSizeTier(
      name,
      SizeTier.small,
      l10n.profileDisplayName,
      l10n,
    );
    if (tooLong != null) {
      setState(() => _nameError = tooLong);
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
    final invitations = _invitations ?? [];

    final isWide = MediaQuery.sizeOf(context).width >= 600;
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380, maxHeight: 600),
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.account_circle, size: 72, color: colorScheme.primary),
              const SizedBox(height: 10),
              if (name != null && name.isNotEmpty)
                Text(
                  name,
                  style: textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
                ),
              if (email != null && email.isNotEmpty) ...[
                const SizedBox(height: 2),
                Text(
                  email,
                  style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant),
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
                      Icon(Icons.language, size: 20, color: colorScheme.onSurfaceVariant),
                      const SizedBox(width: 12),
                      Text(l10n.profileLanguage, style: textTheme.bodyMedium),
                    ],
                  ),
                  const SizedBox(height: 10),
                  SizedBox(
                    width: double.infinity,
                    child: SegmentedButton<String>(
                      segments: const [
                        ButtonSegment(value: 'en', label: Text('English', style: TextStyle(fontSize: 12))),
                        ButtonSegment(value: 'pl', label: Text('Polish', style: TextStyle(fontSize: 12))),
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
                        counterText: SizeTier.small.counterText,
                      ),
                      autofocus: true,
                      maxLength: SizeTier.small.maxLength,
                      textInputAction: TextInputAction.done,
                      onChanged: (v) {
                        setState(() {
                          _nameError = validateSizeTier(
                            v,
                            SizeTier.small,
                            l10n.profileDisplayName,
                            l10n,
                          );
                        });
                      },
                      onSubmitted: (_) => _saveName(l10n),
                    ),
                    const SizedBox(height: 10),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton(
                          onPressed: _savingName ? null : () => setState(() => _editingName = false),
                          child: Text(l10n.generalCancel),
                        ),
                        const SizedBox(width: 8),
                        FilledButton(
                          onPressed: _savingName ? null : () => _saveName(l10n),
                          child: _savingName
                              ? const SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
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
                  leading: Icon(Icons.badge_outlined, size: 20, color: colorScheme.onSurfaceVariant),
                  title: Text(l10n.profileDisplayName, style: textTheme.bodyMedium),
                  subtitle: (name != null && name.isNotEmpty)
                      ? Text(name, style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant))
                      : null,
                  trailing: Icon(Icons.edit_outlined, size: 18, color: colorScheme.onSurfaceVariant),
                  onTap: () => setState(() {
                    _nameController.text = widget.storage.name ?? '';
                    _editingName = true;
                  }),
                ),

              const SizedBox(height: 4),
              const Divider(height: 1),
              const SizedBox(height: 4),

              // Invitations
              if (_invitations == null)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 12),
                  child: Center(child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))),
                )
              else if (invitations.isNotEmpty) ...[
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  child: Row(
                    children: [
                      Icon(Icons.mail_outline, size: 20, color: colorScheme.onSurfaceVariant),
                      const SizedBox(width: 12),
                      Text(l10n.invitationTitle, style: textTheme.bodyMedium),
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 1),
                        decoration: BoxDecoration(
                          color: colorScheme.error,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Text(
                          '${invitations.length}',
                          style: textTheme.labelSmall?.copyWith(color: colorScheme.onError),
                        ),
                      ),
                    ],
                  ),
                ),
                ...invitations.map((inv) => _InvitationRow(
                      invitation: inv,
                      actioning: _actioning.contains(inv.id),
                      onAccept: () => _accept(inv.id),
                      onDecline: () => _decline(inv.id),
                    )),
                const SizedBox(height: 4),
                const Divider(height: 1),
                const SizedBox(height: 4),
              ],

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
                        style: textTheme.bodyMedium?.copyWith(color: colorScheme.error),
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

class _InvitationRow extends StatelessWidget {
  final MyInvitation invitation;
  final bool actioning;
  final VoidCallback onAccept;
  final VoidCallback onDecline;

  const _InvitationRow({
    required this.invitation,
    required this.actioning,
    required this.onAccept,
    required this.onDecline,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(invitation.apiaryName,
                    style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600)),
                Text(
                  invitation.invitedByName,
                  style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant),
                ),
              ],
            ),
          ),
          if (actioning)
            const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2))
          else
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextButton(
                  onPressed: onDecline,
                  style: TextButton.styleFrom(
                    foregroundColor: colorScheme.error,
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    visualDensity: VisualDensity.compact,
                  ),
                  child: Text(l10n.invitationDecline),
                ),
                FilledButton(
                  onPressed: onAccept,
                  style: FilledButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    visualDensity: VisualDensity.compact,
                  ),
                  child: Text(l10n.invitationAccept),
                ),
              ],
            ),
        ],
      ),
    );
  }
}
