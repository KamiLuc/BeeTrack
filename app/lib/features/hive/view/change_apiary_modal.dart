import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../data/hive_repository.dart';

enum _ConfirmState { idle, loading }

class ChangeApiaryModal extends StatefulWidget {
  final int apiaryId;
  final int hiveId;
  final ApiClient apiClient;
  final List<Apiary> otherApiaries;

  const ChangeApiaryModal({
    super.key,
    required this.apiaryId,
    required this.hiveId,
    required this.apiClient,
    required this.otherApiaries,
  });

  @override
  State<ChangeApiaryModal> createState() => _ChangeApiaryModalState();
}

class _ChangeApiaryModalState extends State<ChangeApiaryModal> {
  Apiary? _selected;
  _ConfirmState _confirmState = _ConfirmState.idle;
  String? _errorMessage;

  Future<void> _confirm() async {
    if (_selected == null || _confirmState == _ConfirmState.loading) return;
    setState(() {
      _confirmState = _ConfirmState.loading;
      _errorMessage = null;
    });
    try {
      await HiveRepository(api: context.read<ApiClient>()).changeApiary(
        apiaryId: widget.apiaryId,
        hiveId: widget.hiveId,
        targetApiaryId: _selected!.id,
      );
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context)!;
      setState(() {
        _confirmState = _ConfirmState.idle;
        _errorMessage = (e is ApiException && e.code == 'TARGET_APIARY_FULL')
            ? l10n.hiveChangeApiaryNoSpace
            : l10n.generalError;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final isWide = MediaQuery.sizeOf(context).width >= 600;

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
              Text(l10n.hiveChangeApiaryTitle, style: textTheme.titleMedium),
              const SizedBox(height: 16),
              const Divider(height: 1),
              const SizedBox(height: 4),
              ...widget.otherApiaries.map((apiary) => RadioListTile<Apiary>(
                    value: apiary,
                    groupValue: _selected,
                    dense: true,
                    contentPadding: EdgeInsets.zero,
                    title: Text(apiary.name, style: textTheme.bodyMedium),
                    onChanged: _confirmState == _ConfirmState.loading
                        ? null
                        : (v) => setState(() => _selected = v),
                  )),
              const SizedBox(height: 4),
              const Divider(height: 1),
              if (_errorMessage != null) ...[
                const SizedBox(height: 10),
                Text(
                  _errorMessage!,
                  style:
                      textTheme.bodySmall?.copyWith(color: colorScheme.error),
                  textAlign: TextAlign.center,
                ),
              ],
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: _confirmState == _ConfirmState.loading
                        ? null
                        : () => Navigator.of(context).pop(),
                    child: Text(l10n.generalCancel),
                  ),
                  const SizedBox(width: 8),
                  OutlinedButton(
                    onPressed: (_selected == null ||
                            _confirmState == _ConfirmState.loading)
                        ? null
                        : _confirm,
                    child: _confirmState == _ConfirmState.loading
                        ? SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: colorScheme.primary),
                          )
                        : Text(l10n.generalConfirm),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
