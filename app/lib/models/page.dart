import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class Page {
  final int pageId;
  final int chapterId;
  final String imageUrl;
  final String fileName;
  final String filePath;
  final String status;
  final String? errorMessage;
  final DateTime createdAt;
  final DateTime? downloadedAt;

  Page({
    required this.pageId,
    required this.chapterId,
    required this.imageUrl,
    required this.fileName,
    required this.filePath,
    required this.status,
    this.errorMessage,
    required this.createdAt,
    this.downloadedAt,
  });

  factory Page.fromJson(Map<String, dynamic> json) {
    return Page(
      pageId: json['PageID'] as int,
      chapterId: json['ChapterID'] as int,
      imageUrl: (json['ImageUrl'] as String?) ?? '',
      fileName: (json['FileName'] as String?) ?? 'unknown',
      filePath: json['FilePath'] as String? ?? '',
      status: json['Status'] as String? ?? 'pending',
      errorMessage: json['ErrorMessage'] as String?,
      createdAt: DateTime.parse((json['CreatedAt'] as String?) ?? DateTime.now().toIso8601String()),
      downloadedAt: json['DownloadedAt'] != null
          ? DateTime.parse(json['DownloadedAt'] as String)
          : null,
    );
  }

  // Get full image URL for the page
  String get imageFullUrl {
    if (filePath.isNotEmpty) {
      final host = dotenv.env['IMAGE_HOST']?.trim() ?? dotenv.env['API_HOST']?.trim() ?? 'localhost';
      return 'http://$host:8080$filePath';
    }
    return imageUrl;
  }

  bool get isDownloaded => status.toLowerCase() == 'downloaded';
}
