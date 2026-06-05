import 'package:flutter/widgets.dart';

class AppLayout {
  static BoxConstraints formConstraints(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    return BoxConstraints(maxWidth: width * (width > 900 ? 0.4 : 0.9));
  }
}
