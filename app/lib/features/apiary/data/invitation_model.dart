class ApiaryMemberInfo {
  final int userId;
  final String name;
  final String email;
  final String role;
  final DateTime joinedAt;

  const ApiaryMemberInfo({
    required this.userId,
    required this.name,
    required this.email,
    required this.role,
    required this.joinedAt,
  });

  factory ApiaryMemberInfo.fromJson(Map<String, dynamic> json) => ApiaryMemberInfo(
        userId: json['user_id'] as int,
        name: json['name'] as String,
        email: json['email'] as String,
        role: json['role'] as String,
        joinedAt: DateTime.parse(json['joined_at'] as String).toLocal(),
      );
}

class ApiaryInvitation {
  final int id;
  final int apiaryId;
  final String invitedEmail;
  final String status;
  final DateTime createdAt;

  const ApiaryInvitation({
    required this.id,
    required this.apiaryId,
    required this.invitedEmail,
    required this.status,
    required this.createdAt,
  });

  factory ApiaryInvitation.fromJson(Map<String, dynamic> json) => ApiaryInvitation(
        id: json['id'] as int,
        apiaryId: json['apiary_id'] as int,
        invitedEmail: json['invited_email'] as String,
        status: json['status'] as String,
        createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
      );
}

class MyInvitation {
  final int id;
  final int apiaryId;
  final String apiaryName;
  final String invitedByName;
  final DateTime createdAt;

  const MyInvitation({
    required this.id,
    required this.apiaryId,
    required this.apiaryName,
    required this.invitedByName,
    required this.createdAt,
  });

  factory MyInvitation.fromJson(Map<String, dynamic> json) => MyInvitation(
        id: json['id'] as int,
        apiaryId: json['apiary_id'] as int,
        apiaryName: json['apiary_name'] as String,
        invitedByName: json['invited_by_name'] as String,
        createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
      );
}

class ApiaryMembersData {
  final List<ApiaryMemberInfo> members;
  final List<ApiaryInvitation> invitations;

  const ApiaryMembersData({required this.members, required this.invitations});
}
