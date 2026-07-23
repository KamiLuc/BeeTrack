import 'dart:math';

import 'package:flutter/widgets.dart';

class AppLayout {
  static BoxConstraints formConstraints(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    return BoxConstraints(maxWidth: width * (width > 900 ? 0.4 : 0.9));
  }

  static double bannerWidth(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    return width < 600 ? width * 0.85 : min(440.0, width * 0.40);
  }

  /// Width for a centered modal dialog: most of the screen on phones, capped
  /// well short of the viewport on wide desktop/web windows so it reads as a
  /// card, not a near-full-bleed panel.
  static double dialogWidth(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    return min(560.0, width * 0.9);
  }
}
