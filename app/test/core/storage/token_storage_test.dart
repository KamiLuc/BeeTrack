import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/storage/token_storage.dart';

String _jwtWithPayload(Map<String, dynamic> payload) {
  String segment(Object data) =>
      base64Url.encode(utf8.encode(jsonEncode(data))).replaceAll('=', '');
  final header = segment({'alg': 'none', 'typ': 'JWT'});
  final body = segment(payload);
  return '$header.$body.signature';
}

Future<TokenStorage> _storageWithAccessToken(String? token) async {
  SharedPreferences.setMockInitialValues(
    token == null ? {} : {'access_token': token},
  );
  final prefs = await SharedPreferences.getInstance();
  return TokenStorage(prefs);
}

void main() {
  group('TokenStorage.userId', () {
    test('decodes an integer sub claim from a valid token', () async {
      final storage = await _storageWithAccessToken(_jwtWithPayload({'sub': 7}));

      expect(storage.userId, 7);
    });

    test('decodes a string sub claim from a valid token', () async {
      final storage =
          await _storageWithAccessToken(_jwtWithPayload({'sub': '9'}));

      expect(storage.userId, 9);
    });

    test('returns null when there is no stored access token', () async {
      final storage = await _storageWithAccessToken(null);

      expect(storage.userId, isNull);
    });

    test('returns null for a malformed token with the wrong number of parts',
        () async {
      final storage = await _storageWithAccessToken('not-a-jwt');

      expect(storage.userId, isNull);
    });

    test('returns null when the payload segment is not valid base64/JSON',
        () async {
      final storage = await _storageWithAccessToken('header.@@not-base64@@.sig');

      expect(storage.userId, isNull);
    });

    test('returns null when the sub claim is missing', () async {
      final storage =
          await _storageWithAccessToken(_jwtWithPayload({'email': 'a@b.com'}));

      expect(storage.userId, isNull);
    });
  });
}
