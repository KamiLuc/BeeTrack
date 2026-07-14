import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
import '../../hive/view/apiary_grid_screen.dart';
import '../data/invitation_repository.dart';
import 'apiaries_map_screen.dart';
import 'apiary_members_modal.dart';
import 'create_apiary_screen.dart';
import 'edit_apiary_screen.dart';

class ApiariesScreen extends StatelessWidget {
  final ValueChanged<AppSection> onSelectSection;

  const ApiariesScreen({super.key, required this.onSelectSection});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => ApiariesCubit(
        repo: ApiaryRepository(api: context.read()),
      )..load(),
      child: _ApiariesView(onSelectSection: onSelectSection),
    );
  }
}

class _ApiariesView extends StatelessWidget {
  final ValueChanged<AppSection> onSelectSection;

  const _ApiariesView({required this.onSelectSection});

  Future<void> _openCreate(BuildContext context) async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const CreateApiaryScreen()),
    );
    if (context.mounted) context.read<ApiariesCubit>().load();
  }

  void _openMap(BuildContext context, List<Apiary> apiaries) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ApiariesMapScreen(apiaries: apiaries),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.apiaryTitle),
        actions: [ProfileIconButton(onRefresh: () => context.read<ApiariesCubit>().load())],
      ),
      drawer: AuthenticatedAppDrawer(
        current: AppSection.apiaries,
        onSelect: onSelectSection,
      ),
      body: BlocBuilder<ApiariesCubit, ApiariesState>(
        builder: (context, state) {
          if (state is ApiariesLoading || state is ApiariesInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is ApiariesError) {
            return Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    l10n.generalError,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () => context.read<ApiariesCubit>().load(),
                    style: ElevatedButton.styleFrom(
                      minimumSize: const Size(120, 40),
                    ),
                    child: Text(l10n.generalRetry),
                  ),
                ],
              ),
            );
          }
          if (state is ApiariesLoaded) {
            final hasGps =
                state.apiaries.any((a) => a.lat != null && a.lng != null);
            return Column(
              children: [
                Expanded(
                  child: state.apiaries.isEmpty
                      ? Center(child: Text(l10n.apiaryEmpty))
                      : RefreshIndicator(
                          onRefresh: () =>
                              context.read<ApiariesCubit>().load(),
                          child: Center(
                            child: SingleChildScrollView(
                              physics: const AlwaysScrollableScrollPhysics(),
                              padding: const EdgeInsets.fromLTRB(
                                16,
                                16,
                                16,
                                8,
                              ),
                              child: ConstrainedBox(
                                constraints: AppLayout.formConstraints(
                                  context,
                                ),
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    for (final apiary in state.apiaries) ...[
                                      _ApiaryCard(apiary: apiary),
                                      const SizedBox(height: 8),
                                    ],
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ),
                ),
                _ApiariesBanner(
                  l10n: l10n,
                  hasApiaries: state.apiaries.isNotEmpty,
                  hasGps: hasGps,
                  onAdd: () => _openCreate(context),
                  onMap: () => _openMap(context, state.apiaries),
                ),
              ],
            );
          }
          return const SizedBox.shrink();
        },
      ),
    );
  }
}

class _ApiariesBanner extends StatelessWidget {
  final AppLocalizations l10n;
  final bool hasApiaries;
  final bool hasGps;
  final VoidCallback onAdd;
  final VoidCallback onMap;

  const _ApiariesBanner({
    required this.l10n,
    required this.hasApiaries,
    required this.hasGps,
    required this.onAdd,
    required this.onMap,
  });

  @override
  Widget build(BuildContext context) {
    final bannerWidth = AppLayout.bannerWidth(context);

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Center(
          child: SizedBox(
            width: bannerWidth,
            child: Container(
              decoration: const BoxDecoration(
                color: Colors.amber,
                borderRadius: BorderRadius.only(
                  topLeft: Radius.circular(20),
                  topRight: Radius.circular(20),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black26,
                    blurRadius: 8,
                    offset: Offset(0, -2),
                  ),
                ],
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  IconButton(
                    icon: const Icon(Icons.add),
                    iconSize: 28,
                    tooltip: l10n.apiaryAdd,
                    onPressed: onAdd,
                  ),
                  if (hasApiaries)
                    IconButton(
                      icon: const Icon(Icons.map_outlined),
                      iconSize: 28,
                      tooltip: l10n.apiaryMapTooltip,
                      onPressed: hasGps ? onMap : null,
                    ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ApiaryCard extends StatelessWidget {
  final Apiary apiary;

  const _ApiaryCard({required this.apiary});

  Future<void> _openEdit(BuildContext context) async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => EditApiaryScreen(apiary: apiary)),
    );
    if (context.mounted) context.read<ApiariesCubit>().load();
  }

  Future<void> _copy(BuildContext context) async {
    final l10n = AppLocalizations.of(context)!;
    final defaultName = '${apiary.name} (${l10n.apiaryCopySuffix})';
    final name = await _showCopyApiaryDialog(context, defaultName: defaultName, l10n: l10n);
    if (name == null) return;
    try {
      await ApiaryRepository(api: context.read<ApiClient>()).copyApiary(
        apiary.id,
        name: name,
      );
      if (context.mounted) {
        context.read<ApiariesCubit>().load();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.apiaryCopied)),
        );
      }
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    }
  }

  Future<void> _delete(BuildContext context) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.apiaryDeleteConfirm,
      warning: l10n.apiaryDeleteWarning,
      l10n: l10n,
      withPuzzle: apiary.hiveCount > 0,
    );
    if (!confirmed || !context.mounted) return;
    try {
      await ApiaryRepository(api: context.read<ApiClient>()).deleteApiary(apiary.id);
      if (context.mounted) context.read<ApiariesCubit>().load();
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    }
  }

  Future<void> _leave(BuildContext context) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.leaveApiaryConfirm,
      warning: l10n.leaveApiaryWarning,
      l10n: l10n,
      withPuzzle: true,
      confirmLabel: l10n.leaveApiary,
    );
    if (!confirmed || !context.mounted) return;
    final api = context.read<ApiClient>();
    try {
      await InvitationRepository(api: api).leaveApiary(apiary.id);
      if (context.mounted) context.read<ApiariesCubit>().load();
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isOwner = apiary.userRole == 'owner';
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      child: InkWell(
        onTap: () async {
          await Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => ApiaryGridScreen(apiary: apiary),
            ),
          );
          if (context.mounted) context.read<ApiariesCubit>().load();
        },
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 10, 4, 14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      apiary.name,
                      style: textTheme.titleMedium,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 4),
                  _RoleBadge(isOwner: isOwner, l10n: l10n),
                  _ApiaryMenu(
                    apiary: apiary,
                    isOwner: isOwner,
                    l10n: l10n,
                    onEdit: () => _openEdit(context),
                    onCopy: () => _copy(context),
                    onDelete: () => _delete(context),
                    onLeave: () => _leave(context),
                    onMembers: () {
                      final apiClient = context.read<ApiClient>();
                      final ownerEmail = context.read<TokenStorage>().email;
                      showDialog<void>(
                        context: context,
                        builder: (_) => ApiaryMembersModal(
                          apiaryId: apiary.id,
                          apiClient: apiClient,
                          ownerEmail: ownerEmail,
                        ),
                      );
                    },
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 6,
                children: [
                  _InfoChip(
                    icon: Icons.hive_outlined,
                    label: l10n.hiveCount(apiary.hiveCount),
                  ),
                  _InfoChip(
                    icon: Icons.grid_on_outlined,
                    label: '${apiary.gridRows}×${apiary.gridCols}',
                  ),
                  if (apiary.lat != null)
                    _InfoChip(
                      icon: Icons.location_on_outlined,
                      label: 'GPS',
                      color: Colors.green,
                    ),
                ],
              ),
              if (apiary.lastInspectedAt != null) ...[
                const SizedBox(height: 8),
                Row(
                  children: [
                    Icon(
                      Icons.calendar_today_outlined,
                      size: 13,
                      color: colorScheme.onSurfaceVariant,
                    ),
                    const SizedBox(width: 5),
                    Text(
                      DateFormat(
                        'd MMM yyyy',
                        Localizations.localeOf(context).toString(),
                      ).format(apiary.lastInspectedAt!),
                      style: textTheme.bodySmall?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _RoleBadge extends StatelessWidget {
  final bool isOwner;
  final AppLocalizations l10n;

  const _RoleBadge({required this.isOwner, required this.l10n});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final bgColor = colorScheme.surfaceContainerHighest;
    final fgColor = colorScheme.onSurfaceVariant;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        isOwner ? l10n.roleOwner : l10n.roleMember,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: fgColor,
              fontWeight: FontWeight.w600,
            ),
      ),
    );
  }
}

class _ApiaryMenu extends StatelessWidget {
  final Apiary apiary;
  final bool isOwner;
  final AppLocalizations l10n;
  final VoidCallback onEdit;
  final VoidCallback onCopy;
  final VoidCallback onDelete;
  final VoidCallback onLeave;
  final VoidCallback onMembers;

  const _ApiaryMenu({
    required this.apiary,
    required this.isOwner,
    required this.l10n,
    required this.onEdit,
    required this.onCopy,
    required this.onDelete,
    required this.onLeave,
    required this.onMembers,
  });

  @override
  Widget build(BuildContext context) {
    final errorColor = Theme.of(context).colorScheme.error;

    return PopupMenuButton<_ApiaryAction>(
      icon: const Icon(Icons.more_vert),
      onSelected: (action) {
        switch (action) {
          case _ApiaryAction.members:
            onMembers();
          case _ApiaryAction.edit:
            onEdit();
          case _ApiaryAction.copy:
            onCopy();
          case _ApiaryAction.delete:
            onDelete();
          case _ApiaryAction.leave:
            onLeave();
        }
      },
      itemBuilder: (_) => [
        if (isOwner) ...[
          PopupMenuItem(
            value: _ApiaryAction.members,
            child: ListTile(
              leading: const Icon(Icons.people_outline),
              title: Text(l10n.invitationInvite),
              contentPadding: EdgeInsets.zero,
              visualDensity: VisualDensity.compact,
            ),
          ),
          PopupMenuItem(
            value: _ApiaryAction.edit,
            child: ListTile(
              leading: const Icon(Icons.edit_outlined),
              title: Text(l10n.generalEdit),
              contentPadding: EdgeInsets.zero,
              visualDensity: VisualDensity.compact,
            ),
          ),
        ],
        PopupMenuItem(
          value: _ApiaryAction.copy,
          child: ListTile(
            leading: const Icon(Icons.copy_outlined),
            title: Text(l10n.apiaryCopy),
            contentPadding: EdgeInsets.zero,
            visualDensity: VisualDensity.compact,
          ),
        ),
        if (isOwner)
          PopupMenuItem(
            value: _ApiaryAction.delete,
            child: ListTile(
              leading: Icon(Icons.delete_outline, color: errorColor),
              title: Text(l10n.generalDelete, style: TextStyle(color: errorColor)),
              contentPadding: EdgeInsets.zero,
              visualDensity: VisualDensity.compact,
            ),
          )
        else
          PopupMenuItem(
            value: _ApiaryAction.leave,
            child: ListTile(
              leading: Icon(Icons.exit_to_app, color: errorColor),
              title: Text(l10n.leaveApiary, style: TextStyle(color: errorColor)),
              contentPadding: EdgeInsets.zero,
              visualDensity: VisualDensity.compact,
            ),
          ),
      ],
    );
  }
}

Future<String?> _showCopyApiaryDialog(
  BuildContext context, {
  required String defaultName,
  required AppLocalizations l10n,
}) {
  return showDialog<String>(
    context: context,
    builder: (ctx) => CopyApiaryDialog(defaultName: defaultName, l10n: l10n),
  );
}

class CopyApiaryDialog extends StatefulWidget {
  final String defaultName;
  final AppLocalizations l10n;

  const CopyApiaryDialog({super.key, required this.defaultName, required this.l10n});

  @override
  State<CopyApiaryDialog> createState() => _CopyApiaryDialogState();
}

class _CopyApiaryDialogState extends State<CopyApiaryDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.defaultName);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _confirm() {
    final name = _controller.text.trim();
    if (name.isNotEmpty) Navigator.of(context).pop(name);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = widget.l10n;
    return AlertDialog(
      title: Text(l10n.apiaryCopy),
      content: TextField(
        controller: _controller,
        autofocus: true,
        decoration: InputDecoration(
          labelText: l10n.apiaryCopyNewName,
          border: const OutlineInputBorder(),
        ),
        onSubmitted: (_) => _confirm(),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(null),
          child: Text(l10n.generalCancel),
        ),
        TextButton(
          onPressed: _confirm,
          child: Text(l10n.generalConfirm),
        ),
      ],
    );
  }
}

enum _ApiaryAction { members, edit, copy, delete, leave }

class _InfoChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color? color;

  const _InfoChip({required this.icon, required this.label, this.color});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final effectiveColor = color ?? colorScheme.onSurfaceVariant;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: effectiveColor),
        const SizedBox(width: 4),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: effectiveColor,
              ),
        ),
      ],
    );
  }
}

