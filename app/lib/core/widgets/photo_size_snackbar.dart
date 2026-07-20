import 'package:flutter/material.dart';

/// Shows the "photo too large" error as a tall bottom SnackBar, sized to fully
/// cover the amber save/add-photo bar on the inspection form screen.
void showPhotoTooLargeSnackBar(BuildContext context, String message) {
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
