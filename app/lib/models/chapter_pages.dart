import 'chapter.dart';
import 'page.dart';

class ChapterPages {
  final int chapterId;
  final Chapter chapter;
  final List<Page> pages;

  ChapterPages({
    required this.chapterId,
    required this.chapter,
    required this.pages,
  });

  factory ChapterPages.fromJson(Map<String, dynamic> json) {
    final chapterJson = json['chapter'] as Map<String, dynamic>;
    final pagesJson = json['pages'] as List<dynamic>;

    return ChapterPages(
      chapterId: json['chapter_id'] as int,
      chapter: Chapter.fromJson(chapterJson),
      pages: pagesJson.map((p) => Page.fromJson(p)).toList(),
    );
  }

  int get totalPages => pages.length;
  int get downloadedPages => pages.where((p) => p.isDownloaded).length;
}
