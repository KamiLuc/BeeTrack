import 'package:dio/dio.dart';
import 'package:flutter/material.dart';

import '../../../core/api/api_client.dart';
import '../../../l10n/app_localizations.dart';
import '../data/invitation_model.dart';
import '../data/invitation_repository.dart';

enum _SendState { idle, loading }

class ApiaryMembersModal extends StatefulWidget {
  final int apiaryId;
  final ApiClient apiClient;
  final String? ownerEmail;

  const ApiaryMembersModal({
    super.key,
    required this.apiaryId,
    required this.apiClient,
    this.ownerEmail,
  });

  @override
  State<ApiaryMembersModal> createState() => _ApiaryMembersModalState();
}

class _ApiaryMembersModalState extends State<ApiaryMembersModal> {
  late final InvitationRepository _repo;
  final _emailController = TextEditingController();
  ApiaryMembersData? _data;
  bool _loadingData = true;
  _SendState _sendState = _SendState.idle;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _repo = InvitationRepository(api: widget.apiClient);
    _load();
  }

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loadingData = true);
    try {
      final data = await _repo.listForApiary(widget.apiaryId);
      if (mounted) setState(() { _data = data; _loadingData = false; });
    } catch (_) {
      if (mounted) setState(() => _loadingData = false);
    }
  }

  static final _emailRegex = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');

  Future<void> _send() async {
    if (_sendState == _SendState.loading) return;
    final email = _emailController.text.trim();
    if (!_emailRegex.hasMatch(email)) return;
    if (email.toLowerCase() == widget.ownerEmail?.toLowerCase()) return;
    setState(() { _sendState = _SendState.loading; _errorMessage = null; });
    try {
      await _repo.sendInvitation(widget.apiaryId, email);
      if (!mounted) return;
      _emailController.clear();
      setState(() => _sendState = _SendState.idle);
    } catch (e) {
      if (!mounted) return;
      String? code;
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) code = data['code'] as String?;
      }
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _sendState = _SendState.idle;
        _errorMessage = switch (code) {
          'INVITATION_PENDING' => l10n.invitationAlreadyPending,
          'ALREADY_MEMBER'     => l10n.invitationAlreadyMember,
          'CANNOT_INVITE_SELF' => l10n.invitationCannotInviteSelf,
          'USER_NOT_FOUND'     => l10n.invitationUserNotFound,
          _                    => l10n.generalError,
        };
      });
    }
    _load();
  }

  Future<void> _cancelInvitation(int id) async {
    try { await _repo.cancelInvitation(widget.apiaryId, id); } catch (_) {}
    _load();
  }

  Future<void> _removeMember(int userId) async {
    try { await _repo.removeMember(widget.apiaryId, userId); } catch (_) {}
    _load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final isWide = MediaQuery.sizeOf(context).width >= 600;
    final typedEmail = _emailController.text.trim().toLowerCase();
    final inlineError = _errorMessage ?? () {
      if (typedEmail.isEmpty) return null;
      if (typedEmail == widget.ownerEmail?.toLowerCase()) return l10n.invitationCannotInviteSelf;
      if (_data?.invitations.any((inv) => inv.invitedEmail.toLowerCase() == typedEmail) ?? false) return l10n.invitationAlreadyPending;
      return null;
    }();

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380),
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(l10n.invitationInvite, style: textTheme.titleMedium),
              const SizedBox(height: 16),

              TextField(
                controller: _emailController,
                decoration: InputDecoration(
                  labelText: l10n.invitationEmailHint,
                  border: const OutlineInputBorder(),
                  isDense: true,
                ),
                keyboardType: TextInputType.emailAddress,
                textInputAction: TextInputAction.send,
                onSubmitted: (_) => _send(),
                onChanged: (_) => setState(() => _errorMessage = null),
              ),
              const SizedBox(height: 10),

              Align(
                alignment: Alignment.center,
                child: OutlinedButton(
                  onPressed: (_sendState == _SendState.loading ||
                          !_emailRegex.hasMatch(_emailController.text.trim()) ||
                          _emailController.text.trim().toLowerCase() == widget.ownerEmail?.toLowerCase() ||
                          (_data?.invitations.any((inv) => inv.invitedEmail.toLowerCase() == _emailController.text.trim().toLowerCase()) ?? false))
                      ? null
                      : _send,
                  child: _sendState == _SendState.loading
                      ? SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2, color: colorScheme.primary),
                        )
                      : Text(l10n.invitationSend),
                ),
              ),

              if (inlineError != null) ...[
                const SizedBox(height: 8),
                Text(
                  inlineError,
                  style: textTheme.bodySmall?.copyWith(color: colorScheme.error),
                  textAlign: TextAlign.center,
                ),
              ],

              if (_loadingData || (_data != null && (_data!.members.isNotEmpty || _data!.invitations.isNotEmpty))) ...[
                const SizedBox(height: 20),
                const Divider(height: 1),
                const SizedBox(height: 12),
              ],

              if (_loadingData)
                const Center(
                  child: Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: CircularProgressIndicator(),
                  ),
                )
              else if (_data != null && (_data!.members.isNotEmpty || _data!.invitations.isNotEmpty)) ...[
                if (_data!.members.isNotEmpty) ...[
                  Text(
                    l10n.invitationMembers,
                    style: textTheme.labelSmall?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                      letterSpacing: 0.8,
                    ),
                  ),
                  const SizedBox(height: 4),
                  ..._data!.members.map((m) => ListTile(
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                        leading: Icon(Icons.person_outline, size: 20, color: colorScheme.onSurfaceVariant),
                        title: Text(m.name, style: textTheme.bodyMedium),
                        subtitle: Text(m.email, style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant)),
                        trailing: TextButton(
                          onPressed: () => _removeMember(m.userId),
                          style: TextButton.styleFrom(foregroundColor: colorScheme.error),
                          child: Text(l10n.invitationRemove),
                        ),
                      )),
                ],
                if (_data!.members.isNotEmpty && _data!.invitations.isNotEmpty)
                  const SizedBox(height: 8),
                if (_data!.invitations.isNotEmpty) ...[
                  Text(
                    l10n.invitationPending,
                    style: textTheme.labelSmall?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                      letterSpacing: 0.8,
                    ),
                  ),
                  const SizedBox(height: 4),
                  ..._data!.invitations.map((inv) => ListTile(
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                        leading: Icon(Icons.hourglass_empty, size: 20, color: colorScheme.onSurfaceVariant),
                        title: Text(inv.invitedEmail, style: textTheme.bodyMedium),
                        trailing: TextButton(
                          onPressed: () => _cancelInvitation(inv.id),
                          style: TextButton.styleFrom(foregroundColor: colorScheme.error),
                          child: Text(l10n.invitationRemove),
                        ),
                      )),
                ],
              ],

              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: Text(l10n.generalClose),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
