import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class LocaleController extends ValueNotifier<Locale> {
  static const _key = 'locale_code';

  final SharedPreferences _prefs;

  LocaleController(SharedPreferences prefs)
      : _prefs = prefs,
        super(Locale(prefs.getString(_key) ?? 'pl'));

  Future<void> setLocale(Locale locale) async {
    value = locale;
    await _prefs.setString(_key, locale.languageCode);
  }
}
