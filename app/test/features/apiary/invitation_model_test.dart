import 'package:flutter_test/flutter_test.dart';
import 'package:app/features/apiary/data/invitation_model.dart';

void main() {
  group('ApiaryMemberInfo.fromJson', () {
    test('parses all fields', () {
      final json = {
        'user_id': 42,
        'name': 'Alice',
        'email': 'alice@example.com',
        'role': 'member',
        'joined_at': '2024-01-15T10:00:00Z',
      };
      final m = ApiaryMemberInfo.fromJson(json);
      expect(m.userId, 42);
      expect(m.name, 'Alice');
      expect(m.email, 'alice@example.com');
      expect(m.role, 'member');
      expect(m.joinedAt, DateTime.parse('2024-01-15T10:00:00Z'));
    });
  });

  group('ApiaryInvitation.fromJson', () {
    test('parses all fields', () {
      final json = {
        'id': 7,
        'apiary_id': 3,
        'invited_email': 'bob@example.com',
        'status': 'pending',
        'created_at': '2024-02-01T08:30:00Z',
      };
      final inv = ApiaryInvitation.fromJson(json);
      expect(inv.id, 7);
      expect(inv.apiaryId, 3);
      expect(inv.invitedEmail, 'bob@example.com');
      expect(inv.status, 'pending');
      expect(inv.createdAt, DateTime.parse('2024-02-01T08:30:00Z'));
    });
  });

  group('MyInvitation.fromJson', () {
    test('parses all fields', () {
      final json = {
        'id': 12,
        'apiary_id': 5,
        'apiary_name': 'Home Apiary',
        'invited_by_name': 'Charlie',
        'created_at': '2024-03-10T12:00:00Z',
      };
      final inv = MyInvitation.fromJson(json);
      expect(inv.id, 12);
      expect(inv.apiaryId, 5);
      expect(inv.apiaryName, 'Home Apiary');
      expect(inv.invitedByName, 'Charlie');
      expect(inv.createdAt, DateTime.parse('2024-03-10T12:00:00Z'));
    });
  });

  group('ApiaryMembersData', () {
    test('holds members and invitations', () {
      final data = ApiaryMembersData(
        members: [
          ApiaryMemberInfo(
            userId: 1,
            name: 'Alice',
            email: 'alice@example.com',
            role: 'member',
            joinedAt: DateTime(2024),
          ),
        ],
        invitations: [
          ApiaryInvitation(
            id: 1,
            apiaryId: 1,
            invitedEmail: 'bob@example.com',
            status: 'pending',
            createdAt: DateTime(2024),
          ),
        ],
      );
      expect(data.members.length, 1);
      expect(data.invitations.length, 1);
    });
  });
}
