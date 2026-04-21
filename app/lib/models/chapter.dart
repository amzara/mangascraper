class Chapter {
  final int chapterId;
  final int mangaId;
  final String chapterSlug;
  final double chapterNum;
  final String status;
  final String? errorMessage;
  final DateTime createdAt;
  final DateTime? updatedAt;

  Chapter({
    required this.chapterId,
    required this.mangaId,
    required this.chapterSlug,
    required this.chapterNum,
    required this.status,
    this.errorMessage,
    required this.createdAt,
    this.updatedAt,
  });

  factory Chapter.fromJson(Map<String, dynamic> json) {
    return Chapter(
      chapterId: json['ChapterID'] as int,
      mangaId: json['MangaID'] as int,
      chapterSlug: json['ChapterSlug'] as String,
      chapterNum: (json['ChapterNum'] as num?)?.toDouble() ?? 0.0,
      status: json['Status'] as String,
      errorMessage: json['ErrorMessage'] as String?,
      createdAt: DateTime.parse(json['CreatedAt'] as String),
      updatedAt: json['UpdatedAt'] != null
          ? DateTime.parse(json['UpdatedAt'] as String)
          : null,
    );
  }

  // Get display name (e.g., "Chapter 1")
  String get displayName {
    if (chapterNum > 0) {
      return 'Chapter $chapterNum';
    }
    // Fallback to slug parsing
    return chapterSlug
        .replaceAll('-', ' ')
        .split(' ')
        .map((word) => word.isNotEmpty
            ? '${word[0].toUpperCase()}${word.substring(1)}'
            : '')
        .join(' ');
  }

  // Get status color
  int get statusColor {
    switch (status.toLowerCase()) {
      case 'downloaded':
        return 0xFF4CAF50; // Green
      case 'downloading':
        return 0xFF2196F3; // Blue
      case 'error':
        return 0xFFF44336; // Red
      case 'pending':
      default:
        return 0xFFFF9800; // Orange
    }
  }
}
