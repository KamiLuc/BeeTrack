import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../l10n/app_localizations.dart';
import '../../auth/bloc/auth_bloc.dart';
import '../cubit/apiaries_cubit.dart';
import '../data/apiary_model.dart';
import '../data/apiary_repository.dart';
import 'create_apiary_screen.dart';

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
              return Center(child: Text(l10n.apiaryEmpty));
            }
            return RefreshIndicator(
              onRefresh: () => context.read<ApiariesCubit>().load(),
              child: ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: state.apiaries.length,
                separatorBuilder: (_, _) => const SizedBox(height: 8),
                itemBuilder: (context, i) =>
                    _ApiaryCard(apiary: state.apiaries[i]),
              ),
            );
          }
          return const SizedBox.shrink();
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          await Navigator.of(context).push(
            MaterialPageRoute(builder: (_) => const CreateApiaryScreen()),
          );
          if (context.mounted) context.read<ApiariesCubit>().load();
        },
        child: const Icon(Icons.add),
      ),
    );
  }
}

class _ApiaryCard extends StatelessWidget {
  final Apiary apiary;

  const _ApiaryCard({required this.apiary});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        title: Text(apiary.name),
        subtitle: apiary.lat != null
            ? Text('${apiary.lat!.toStringAsFixed(4)}, ${apiary.lng!.toStringAsFixed(4)}')
            : null,
        trailing: Chip(
          label: Text(apiary.userRole),
          labelStyle: Theme.of(context).textTheme.labelSmall,
        ),
        onTap: () {},
      ),
    );
  }
}
