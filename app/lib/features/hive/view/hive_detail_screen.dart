import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import 'edit_hive_screen.dart';
import 'hive_form_widgets.dart';

class HiveDetailScreen extends StatefulWidget {
  final Hive hive;
  final int apiaryId;

  const HiveDetailScreen({
    super.key,
    required this.hive,
    required this.apiaryId,
  });

  @override
  State<HiveDetailScreen> createState() => _HiveDetailScreenState();
}

class _HiveDetailScreenState extends State<HiveDetailScreen> {
  late Hive _hive;

  @override
  void initState() {
    super.initState();
    _hive = widget.hive;
  }

  Future<void> _openEdit() async {
    final updated = await Navigator.of(context).push<Hive>(
      MaterialPageRoute(
        builder: (_) => EditHiveScreen(apiaryId: widget.apiaryId, hive: _hive),
      ),
    );
    if (updated != null && mounted) {
      setState(() => _hive = updated);
    }
  }

  Future<void> _confirmDelete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.hiveDeleteConfirm),
        content: Text(l10n.hiveDeleteWarning),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(l10n.generalCancel),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: Text(
              l10n.generalDelete,
              style: TextStyle(color: Theme.of(ctx).colorScheme.error),
            ),
          ),
        ],
      ),
    );
    if (!(confirmed ?? false) || !mounted) return;
    await HiveRepository(api: context.read()).deleteHive(
      apiaryId: widget.apiaryId,
      hiveId: _hive.id,
    );
    if (mounted) Navigator.of(context).pop(true);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: Text(_hive.name),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            onPressed: _openEdit,
          ),
          PopupMenuButton<_DetailAction>(
            onSelected: (action) {
              if (action == _DetailAction.delete) _confirmDelete();
            },
            itemBuilder: (_) => [
              PopupMenuItem(
                value: _DetailAction.delete,
                child: Text(
                  l10n.generalDelete,
                  style: TextStyle(color: colorScheme.error),
                ),
              ),
            ],
          ),
        ],
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: ConstrainedBox(
            constraints: AppLayout.formConstraints(context),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _InfoCard(hive: _hive),
                const SizedBox(height: 16),
                _SectionCard(
                  title: l10n.hiveDetailInspections,
                  icon: Icons.search,
                  emptyText: l10n.hiveDetailNoInspections,
                  actionLabel: l10n.hiveDetailAddInspection,
                  onAction: null,
                ),
                const SizedBox(height: 16),
                _SectionCard(
                  title: l10n.hiveDetailTreatments,
                  icon: Icons.medical_services_outlined,
                  emptyText: l10n.hiveDetailNoTreatments,
                  actionLabel: l10n.hiveDetailLogTreatment,
                  onAction: null,
                ),
                const SizedBox(height: 16),
                _SectionCard(
                  title: l10n.hiveDetailHarvests,
                  icon: Icons.water_drop_outlined,
                  emptyText: l10n.hiveDetailNoHarvests,
                  actionLabel: l10n.hiveDetailLogHarvest,
                  onAction: null,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

enum _DetailAction { delete }

class _InfoCard extends StatelessWidget {
  final Hive hive;

  const _InfoCard({required this.hive});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          children: [
            _InfoRow(
              label: l10n.hiveType,
              value: hiveTypeLabels[hive.type] ?? hive.type,
            ),
            const Divider(height: 24),
            _InfoRow(
              label: l10n.hiveStatus,
              trailing: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 8,
                    height: 8,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: hive.active ? Colors.green : colorScheme.outline,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    hive.active ? l10n.hiveActive : l10n.hiveInactive,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String? value;
  final Widget? trailing;

  const _InfoRow({required this.label, this.value, this.trailing});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: textTheme.bodyMedium?.copyWith(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        if (trailing != null)
          trailing!
        else
          Text(
            value ?? '',
            style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
          ),
      ],
    );
  }
}

class _SectionCard extends StatelessWidget {
  final String title;
  final IconData icon;
  final String emptyText;
  final String actionLabel;
  final VoidCallback? onAction;

  const _SectionCard({
    required this.title,
    required this.icon,
    required this.emptyText,
    required this.actionLabel,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: colorScheme.primary),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(title, style: textTheme.titleMedium),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              emptyText,
              style: textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: onAction,
              child: Text(actionLabel),
            ),
          ],
        ),
      ),
    );
  }
}
