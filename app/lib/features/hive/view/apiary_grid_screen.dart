import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../cubit/hives_cubit.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';

class ApiaryGridScreen extends StatelessWidget {
  final Apiary apiary;

  const ApiaryGridScreen({super.key, required this.apiary});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => HivesCubit(
        repo: HiveRepository(api: context.read()),
        apiaryId: apiary.id,
      )..load(),
      child: _ApiaryGridView(apiary: apiary),
    );
  }
}

class _ApiaryGridView extends StatelessWidget {
  final Apiary apiary;

  const _ApiaryGridView({required this.apiary});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(title: Text(apiary.name)),
      body: BlocBuilder<HivesCubit, HivesState>(
        builder: (context, state) {
          if (state is HivesInitial || state is HivesLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is HivesError) {
            return Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(l10n.generalError),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () => context.read<HivesCubit>().load(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          if (state is HivesLoaded) {
            final hiveMap = {
              for (final h in state.hives) (h.gridRow, h.gridCol): h,
            };
            const double cellSize = 80;
            const double spacing = 8;
            const double padding = 16;
            final maxGridWidth = apiary.gridCols * cellSize +
                (apiary.gridCols - 1) * spacing +
                padding * 2;
            return RefreshIndicator(
              onRefresh: () => context.read<HivesCubit>().load(),
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(padding),
                  child: ConstrainedBox(
                    constraints: BoxConstraints(maxWidth: maxGridWidth),
                    child: GridView.builder(
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: apiary.gridCols,
                        crossAxisSpacing: spacing,
                        mainAxisSpacing: spacing,
                      ),
                      itemCount: apiary.gridRows * apiary.gridCols,
                      itemBuilder: (context, index) {
                        final row = index ~/ apiary.gridCols;
                        final col = index % apiary.gridCols;
                        final hive = hiveMap[(row, col)];
                        return hive != null ? _HiveCell(hive: hive) : const _EmptyCell();
                      },
                    ),
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

class _HiveCell extends StatelessWidget {
  final Hive hive;

  const _HiveCell({required this.hive});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      margin: EdgeInsets.zero,
      child: InkWell(
        onTap: () {},
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.hive,
                size: 28,
                color: hive.active ? colorScheme.primary : colorScheme.outline,
              ),
              const SizedBox(height: 4),
              Text(
                hive.name,
                style: Theme.of(context).textTheme.bodySmall,
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              if (!hive.active)
                Text(
                  l10n.hiveInactive,
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: colorScheme.outline,
                      ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _EmptyCell extends StatelessWidget {
  const _EmptyCell();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Theme.of(context).colorScheme.outlineVariant,
        ),
      ),
    );
  }
}
