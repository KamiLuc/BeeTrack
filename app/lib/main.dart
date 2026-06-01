import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'core/api/api_client.dart';
import 'core/storage/token_storage.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final prefs = await SharedPreferences.getInstance();
  final storage = TokenStorage(prefs);
  final apiClient = ApiClient(
    storage: storage,
    baseUrl: 'http://10.0.2.2:8080',
  );

  runApp(
    MultiRepositoryProvider(
      providers: [
        RepositoryProvider.value(value: storage),
        RepositoryProvider.value(value: apiClient),
      ],
      child: const BeeTrackApp(),
    ),
  );
}

class BeeTrackApp extends StatelessWidget {
  const BeeTrackApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'BeeTrack',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.amber),
        useMaterial3: true,
      ),
      home: const Scaffold(
        body: Center(child: Text('BeeTrack')),
      ),
    );
  }
}
