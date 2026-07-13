import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

class TokenStorage {
  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';
  static const _emailKey = 'user_email';
  static const _nameKey = 'user_name';

  final SharedPreferences _prefs;

  TokenStorage(this._prefs);

  String? get accessToken => _prefs.getString(_accessKey);
  String? get refreshToken => _prefs.getString(_refreshKey);
  String? get email => _prefs.getString(_emailKey);
  String? get name => _prefs.getString(_nameKey);

  /// The `sub` claim of the access token, i.e. the logged-in user's id.
  /// Decoded locally from the (already-verified-by-the-server) JWT payload.
  int? get userId {
    final token = accessToken;
    if (token == null) return null;
    final parts = token.split('.');
    if (parts.length != 3) return null;
    try {
      final payload =
          utf8.decode(base64Url.decode(base64Url.normalize(parts[1])));
      final sub = (jsonDecode(payload) as Map<String, dynamic>)['sub'];
      if (sub is int) return sub;
      if (sub is double) return sub.toInt();
      if (sub is String) return int.tryParse(sub);
      return null;
    } catch (_) {
      return null;
    }
  }

  Future<void> save({
    required String access,
    required String refresh,
    String? email,
    String? name,
  }) async {
    await _prefs.setString(_accessKey, access);
    await _prefs.setString(_refreshKey, refresh);
    if (email != null) await _prefs.setString(_emailKey, email);
    if (name != null) await _prefs.setString(_nameKey, name);
  }

  Future<void> saveName(String name) =>
      _prefs.setString(_nameKey, name);

  Future<void> clear() async {
    await _prefs.remove(_accessKey);
    await _prefs.remove(_refreshKey);
    await _prefs.remove(_emailKey);
    await _prefs.remove(_nameKey);
  }
}
