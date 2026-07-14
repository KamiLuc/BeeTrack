import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/locale/locale_controller.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/core/widgets/profile_icon_button.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/l10n/app_localizations.dart';

/// Serves a canned zero-count response for the invitations badge fetch, so
/// tests don't hang waiting on a real network call.
class _FakeAdapter implements HttpClientAdapter {
  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final body = options.path.contains('/invitations/count')
        ? jsonEncode({'count': 0})
        : jsonEncode([]);
    return ResponseBody.fromString(
      body,
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }
}

Future<Widget> _wrapped() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final storage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: storage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FakeAdapter();
  final localeController = LocaleController(prefs);
  final authBloc = AuthBloc(
    auth: AuthRepository(api: apiClient, storage: storage),
  );

  return MultiProvider(
    providers: [
      RepositoryProvider<ApiClient>.value(value: apiClient),
      RepositoryProvider<TokenStorage>.value(value: storage),
      ChangeNotifierProvider<LocaleController>.value(value: localeController),
    ],
    child: BlocProvider<AuthBloc>.value(
      value: authBloc,
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Scaffold(
          appBar: AppBar(actions: const [ProfileIconButton()]),
        ),
      ),
    ),
  );
}

void main() {
  group('ProfileIconButton display name validation', () {
    testWidgets('truncates display name input at 50 characters',
        (tester) async {
      await tester.pumpWidget(await _wrapped());
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.account_circle_outlined));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Display name'));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'a' * 60);
      await tester.pump();

      final field = tester.widget<TextField>(find.byType(TextField));
      expect(field.controller!.text.length, 50);
    });
  });
}
