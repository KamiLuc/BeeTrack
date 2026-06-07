import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ThemeController extends ValueNotifier<ThemeMode> {
  static const _key = 'theme_mode';

  final SharedPreferences _prefs;

  ThemeController(SharedPreferences prefs)
      : _prefs = prefs,
        super(_load(prefs));

  static ThemeMode _load(SharedPreferences prefs) {
    switch (prefs.getString(_key)) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }

  Future<void> setThemeMode(ThemeMode mode) async {
    value = mode;
    await _prefs.setString(_key, mode.name);
  }
}
