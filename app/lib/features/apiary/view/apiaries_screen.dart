import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
import '../../hive/view/apiary_grid_screen.dart';
import 'apiaries_map_screen.dart';
import 'create_apiary_screen.dart';
import 'edit_apiary_screen.dart';

class ApiariesScreen extends StatelessWidget {
  const ApiariesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => ApiariesCubit(
        repo: ApiaryRepository(api: context.read()),
      )..load(),
      child: const _ApiariesView(),
    );
  }
}

class _ApiariesView extends StatelessWidget {
  const _ApiariesView();

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
        actions: const [ProfileIconButton()],
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
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          if (state is ApiariesLoaded) {
            if (state.apiaries.isEmpty) {
              return Center(
                child: SizedBox(
                  width: 200,
                  child: ElevatedButton(
                    onPressed: () => _openCreate(context),
                    child: Text(l10n.apiaryAdd),
                  ),
                ),
              );
            }
            final hasGps =
                state.apiaries.any((a) => a.lat != null && a.lng != null);
            return Column(
              children: [
                Expanded(
                  child: RefreshIndicator(
                    onRefresh: () => context.read<ApiariesCubit>().load(),
                    child: Center(
                      child: SingleChildScrollView(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                        child: ConstrainedBox(
                          constraints: AppLayout.formConstraints(context),
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
  final bool hasGps;
  final VoidCallback onAdd;
  final VoidCallback onMap;

  const _ApiariesBanner({
    required this.l10n,
    required this.hasGps,
    required this.onAdd,
    required this.onMap,
  });

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isSmallScreen = screenWidth < 600;
    final bannerWidth = isSmallScreen
        ? screenWidth * 0.85
        : min(440.0, screenWidth * 0.40);

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

  Future<void> _confirmDelete(BuildContext context, AppLocalizations l10n) async {
    final cubit = context.read<ApiariesCubit>();
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.apiaryDeleteConfirm,
      warning: l10n.apiaryDeleteWarning,
      l10n: l10n,
      withPuzzle: apiary.hiveCount > 0,
    );
    if (confirmed) {
      cubit.delete(apiary.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isOwner = apiary.userRole == 'owner';

    return SizedBox(
      height: 48,
      child: Card(
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
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(apiary.name, style: Theme.of(context).textTheme.bodyMedium),
                      if (apiary.hiveCount > 0)
                        Text(
                          l10n.hiveCount(apiary.hiveCount),
                          style: Theme.of(context).textTheme.bodySmall,
                        ),
                    ],
                  ),
                ),
                if (apiary.lastInspectedAt != null)
                  Text(
                    DateFormat('d MMM yyyy',
                            Localizations.localeOf(context).toString())
                        .format(apiary.lastInspectedAt!),
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(context).colorScheme.onSurfaceVariant,
                        ),
                  ),
                const SizedBox(width: 4),
                if (isOwner)
                  SizedBox(
                    width: 32,
                    height: 32,
                    child: PopupMenuButton<_ApiaryAction>(
                      padding: EdgeInsets.zero,
                      icon: const Icon(Icons.more_vert, size: 18),
                      onSelected: (action) {
                        if (action == _ApiaryAction.edit) _openEdit(context);
                        if (action == _ApiaryAction.delete) _confirmDelete(context, l10n);
                      },
                      itemBuilder: (_) => [
                        PopupMenuItem(
                          value: _ApiaryAction.edit,
                          child: Text(l10n.generalEdit),
                        ),
                        PopupMenuItem(
                          value: _ApiaryAction.delete,
                          child: Text(
                            l10n.generalDelete,
                            style: TextStyle(color: Theme.of(context).colorScheme.error),
                          ),
                        ),
                      ],
                    ),
                  )
                else
                  Chip(
                    label: Text(l10n.roleMember),
                    labelStyle: Theme.of(context).textTheme.labelSmall,
                    padding: EdgeInsets.zero,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

enum _ApiaryAction { edit, delete }
