import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.apiaryTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: l10n.authLogout,
            onPressed: () =>
                context.read<AuthBloc>().add(LogoutRequested()),
          ),
        ],
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
                  Text(l10n.generalError),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () =>
                        context.read<ApiariesCubit>().load(),
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
            return RefreshIndicator(
              onRefresh: () => context.read<ApiariesCubit>().load(),
              child: Align(
                alignment: Alignment.topCenter,
                child: ConstrainedBox(
                  constraints: AppLayout.formConstraints(context),
                  child: ListView.separated(
                    padding: const EdgeInsets.all(16),
                    itemCount: state.apiaries.length + 1,
                    separatorBuilder: (_, _) => const SizedBox(height: 8),
                    itemBuilder: (context, i) {
                      if (i == state.apiaries.length) {
                        return Padding(
                          padding: const EdgeInsets.only(top: 8),
                          child: Center(
                            child: SizedBox(
                              width: 200,
                              child: ElevatedButton(
                                onPressed: () => _openCreate(context),
                                child: Text(l10n.apiaryAdd),
                              ),
                            ),
                          ),
                        );
                      }
                      return _ApiaryCard(apiary: state.apiaries[i]);
                    },
                  ),
                ),
              ),
            );
          }
          return const SizedBox.shrink();
        },
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
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.apiaryDeleteConfirm),
        content: Text(l10n.apiaryDeleteWarning),
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
    if ((confirmed ?? false) && context.mounted) {
      context.read<ApiariesCubit>().delete(apiary.id);
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
        child: ListTile(
          dense: true,
          visualDensity: VisualDensity.compact,
          contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
          title: Text(apiary.name, style: Theme.of(context).textTheme.bodyMedium),
          subtitle: apiary.hiveCount > 0
              ? Text(
                  '${apiary.hiveCount} ${l10n.hiveTitle.toLowerCase()}',
                  style: Theme.of(context).textTheme.bodySmall,
                )
              : null,
          trailing: isOwner
              ? PopupMenuButton<_ApiaryAction>(
                  padding: EdgeInsets.zero,
                  icon: const Icon(Icons.more_vert, size: 20),
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
                )
              : Chip(
                  label: Text(l10n.roleMember),
                  labelStyle: Theme.of(context).textTheme.labelSmall,
                  padding: EdgeInsets.zero,
                  materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
          onTap: () {},
        ),
      ),
    );
  }
}

enum _ApiaryAction { edit, delete }
