import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'core/api/api_client.dart';
import 'core/locale/locale_controller.dart';
import 'core/storage/token_storage.dart';
import 'core/theme/app_theme.dart';
import 'core/widgets/app_drawer.dart';
import 'features/auth/bloc/auth_bloc.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/auth/view/login_screen.dart';
import 'features/apiary/view/apiaries_screen.dart';
import 'features/honey_batch/view/honey_batches_home_screen.dart';
import 'features/marketplace/view/marketplace_home_screen.dart';
import 'l10n/app_localizations.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final prefs = await SharedPreferences.getInstance();
  final storage = TokenStorage(prefs);
  final apiClient = ApiClient(
    storage: storage,
    baseUrl: kIsWeb ? 'http://localhost:8080' : 'https://beetrack.duckdns.org',
  );
  final localeController = LocaleController(prefs);

  runApp(
    MultiProvider(
      providers: [
        RepositoryProvider.value(value: storage),
        RepositoryProvider.value(value: apiClient),
        ChangeNotifierProvider.value(value: localeController),
      ],
      child: const BeeTrackApp(),
    ),
  );
}

class BeeTrackApp extends StatelessWidget {
  const BeeTrackApp({super.key});

  @override
  Widget build(BuildContext context) {
    final localeController = context.read<LocaleController>();
    return BlocProvider(
      create: (_) => AuthBloc(
        auth: AuthRepository(api: context.read(), storage: context.read()),
      )..add(AppStarted()),
      child: ValueListenableBuilder<Locale>(
        valueListenable: localeController,
        builder: (context, locale, _) => MaterialApp(
          title: 'BeeTrack',
          theme: AppTheme.light(),
          locale: locale,
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          home: const AuthWrapper(),
        ),
      ),
    );
  }
}

class AuthWrapper extends StatefulWidget {
  const AuthWrapper({super.key});

  @override
  State<AuthWrapper> createState() => _AuthWrapperState();
}

class _AuthWrapperState extends State<AuthWrapper> {
  AppSection _section = AppSection.apiaries;
  // Apiaries is the app's default section. Since it requires auth, an
  // unauthenticated user lands on the login gate by default, not Marketplace
  // — Marketplace is reached only via explicit drawer navigation.
  bool _showLogin = true;

  @override
  Widget build(BuildContext context) {
    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthAuthenticated) {
          setState(() {
            _section = AppSection.apiaries;
            _showLogin = false;
          });
        } else if (state is AuthUnauthenticated) {
          Navigator.of(context).popUntil((route) => route.isFirst);
          setState(() => _showLogin = true);
        }
      },
      child: BlocBuilder<AuthBloc, AuthState>(
        builder: (context, state) {
          if (state is AuthAuthenticated) {
            return switch (_section) {
              AppSection.apiaries => ApiariesScreen(
                  onSelectSection: (s) => setState(() => _section = s),
                ),
              AppSection.marketplace => MarketplaceHomeScreen(
                  onSelectSection: (s) => setState(() => _section = s),
                ),
              AppSection.honeyBatches => HoneyBatchesHomeScreen(
                  onSelectSection: (s) => setState(() => _section = s),
                ),
            };
          }
          if (_showLogin) {
            return LoginScreen(
              drawer: UnauthenticatedAppDrawer(
                isLogin: true,
                onMarketplace: () => setState(() => _showLogin = false),
                onLogin: () {},
              ),
            );
          }
          return MarketplaceHomeScreen(
            onLogin: () => setState(() => _showLogin = true),
          );
        },
      ),
    );
  }
}
