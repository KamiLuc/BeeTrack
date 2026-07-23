import 'package:flutter/material.dart';

/// Shows message as a tall bottom SnackBar, sized to fully cover an amber
/// save/add-photo bar rather than the default thin SnackBar — used
/// consistently for form-level errors (oversized files, GPS unavailable,
/// generic save failures) so they all read at the same size.
void showBigSnackBar(BuildContext context, String message) {
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      behavior: SnackBarBehavior.fixed,
      content: SizedBox(
        height: 48,
        child: Center(child: Text(message)),
      ),
    ),
  );
}
