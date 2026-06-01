import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../l10n/app_localizations.dart';
import '../../auth/bloc/auth_bloc.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: const Text('BeeTrack'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: l10n.authLogout,
            onPressed: () =>
                context.read<AuthBloc>().add(LogoutRequested()),
          ),
        ],
      ),
      body: const Center(child: Text('BeeTrack')),
    );
  }
}
