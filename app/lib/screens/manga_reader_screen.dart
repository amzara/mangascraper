import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:visibility_detector/visibility_detector.dart';
import '../models/chapter.dart';
import '../models/chapter_pages.dart';
import '../models/page.dart' as manga_page;
import '../services/api_service.dart';

class MangaReaderScreen extends StatefulWidget {
  final Chapter chapter;
  final String mangaTitle;

  const MangaReaderScreen({
    super.key,
    required this.chapter,
    required this.mangaTitle,
  });

  @override
  State<MangaReaderScreen> createState() => _MangaReaderScreenState();
}

class _MangaReaderScreenState extends State<MangaReaderScreen> {
  final ApiService _apiService = ApiService();
  ChapterPages? _chapterPages;
  bool _isLoading = true;
  String? _error;
  int _currentPage = 0;
  final ScrollController _scrollController = ScrollController();
  
  // Track which pages are visible to manage memory
  final Set<int> _visiblePages = {};
  final Set<int> _preloadedPages = {};

  @override
  void initState() {
    super.initState();
    _loadChapterPages();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _loadChapterPages() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final pages = await _apiService.getChapterPages(widget.chapter.chapterId);

      setState(() {
        _chapterPages = pages;
        _isLoading = false;
      });
      
      // Preload first few images
      _preloadAdjacentPages(0);
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  void _onVisibilityChanged(int pageIndex, VisibilityInfo info) {
    setState(() {
      if (info.visibleFraction > 0.1) {
        _visiblePages.add(pageIndex);
        // Preload adjacent pages when this page becomes visible
        _preloadAdjacentPages(pageIndex);
      } else {
        _visiblePages.remove(pageIndex);
      }
    });
  }

  void _preloadAdjacentPages(int currentIndex) {
    if (_chapterPages == null) return;
    
    // Preload next 3 and previous 1 pages
    final pagesToPreload = [
      currentIndex - 1,
      currentIndex + 1,
      currentIndex + 2,
      currentIndex + 3,
    ];
    
    for (final index in pagesToPreload) {
      if (index >= 0 && 
          index < _chapterPages!.pages.length && 
          !_preloadedPages.contains(index)) {
        _preloadedPages.add(index);
        final page = _chapterPages!.pages[index];
        
        // Preload using CachedNetworkImage provider
        final provider = CachedNetworkImageProvider(
          page.imageFullUrl,
          maxWidth: 1200,
        );
        // This triggers the load
        provider.resolve(const ImageConfiguration());
      }
    }
  }

  void _onScroll() {
    if (_chapterPages == null || _chapterPages!.pages.isEmpty) return;
    
    // Calculate current page based on scroll position
    final scrollOffset = _scrollController.offset;
    final screenHeight = MediaQuery.of(context).size.height;
    final estimatedPageHeight = screenHeight * 1.4;
    final currentIndex = (scrollOffset / estimatedPageHeight).round();
    final newPage = currentIndex.clamp(0, _chapterPages!.pages.length - 1);
    
    if (newPage != _currentPage) {
      setState(() {
        _currentPage = newPage;
      });
      // Preload adjacent pages when scrolling to new page
      _preloadAdjacentPages(newPage);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black87,
        foregroundColor: Colors.white,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.mangaTitle,
              style: const TextStyle(fontSize: 14),
            ),
            Text(
              widget.chapter.displayName,
              style: const TextStyle(fontSize: 12, color: Colors.grey),
            ),
          ],
        ),
        actions: [
          if (_chapterPages != null)
            Padding(
              padding: const EdgeInsets.only(right: 16),
              child: Center(
                child: Text(
                  'Page ${_currentPage + 1}/${_chapterPages!.pages.length}',
                  style: const TextStyle(color: Colors.white),
                ),
              ),
            ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: Colors.white),
      );
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'Error: $_error',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.white),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadChapterPages,
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (_chapterPages == null || _chapterPages!.pages.isEmpty) {
      return const Center(
        child: Text(
          'No pages found',
          style: TextStyle(color: Colors.white),
        ),
      );
    }

    return NotificationListener<ScrollNotification>(
      onNotification: (notification) {
        if (notification is ScrollUpdateNotification) {
          _onScroll();
        }
        return false;
      },
      child: ListView.builder(
        controller: _scrollController,
        cacheExtent: 0, // Don't cache off-screen widgets
        itemCount: _chapterPages!.pages.length,
        itemBuilder: (context, index) {
          final page = _chapterPages!.pages[index];
          final isVisible = _visiblePages.contains(index);
          
          return VisibilityDetector(
            key: Key('page_$index'),
            onVisibilityChanged: (info) => _onVisibilityChanged(index, info),
            child: _MangaPageImage(
              page: page,
              isVisible: isVisible,
            ),
          );
        },
      ),
    );
  }
}

class _MangaPageImage extends StatelessWidget {
  final manga_page.Page page;
  final bool isVisible;

  const _MangaPageImage({
    required this.page,
    required this.isVisible,
  });

  @override
  Widget build(BuildContext context) {
    // Get screen width for cache width calculation
    final screenWidth = MediaQuery.of(context).size.width;
    final cacheWidth = (screenWidth * 1.5).toInt(); // 1.5x for quality
    
    return InteractiveViewer(
      minScale: 0.5,
      maxScale: 4.0,
      child: Container(
        width: double.infinity,
        color: Colors.black,
        child: CachedNetworkImage(
          imageUrl: page.imageFullUrl,
          memCacheWidth: cacheWidth,
          maxWidthDiskCache: cacheWidth,
          fadeInDuration: Duration.zero, // Instant display
          fadeOutDuration: Duration.zero,
          placeholder: (context, url) => Container(
            height: MediaQuery.of(context).size.width * 1.4,
            color: Colors.grey[900],
            child: const Center(
              child: CircularProgressIndicator(color: Colors.white54),
            ),
          ),
          errorWidget: (context, url, error) => Container(
            height: MediaQuery.of(context).size.width * 1.4,
            color: Colors.grey[900],
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(
                  Icons.error_outline,
                  color: Colors.red,
                  size: 48,
                ),
                const SizedBox(height: 8),
                Text(
                  'Failed to load\n${page.fileName}',
                  textAlign: TextAlign.center,
                  style: const TextStyle(color: Colors.white),
                ),
                if (!page.isDownloaded)
                  const Padding(
                    padding: EdgeInsets.only(top: 8),
                    child: Text(
                      'Image not yet downloaded',
                      style: TextStyle(color: Colors.orange, fontSize: 12),
                    ),
                  ),
              ],
            ),
          ),
          imageBuilder: (context, imageProvider) {
            return Image(
              image: imageProvider,
              fit: BoxFit.fitWidth,
              width: double.infinity,
              filterQuality: FilterQuality.medium,
            );
          },
        ),
      ),
    );
  }
}
