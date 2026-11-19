#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

extern void peregrineAppHidden(int hidden);
extern void peregrineReopen(void);

// Dock click (or opening the app again from Finder): if no window is on screen, show it.
static BOOL peregrineShouldHandleReopen(id self, SEL _cmd, NSApplication *sender, BOOL hasVisibleWindows) {
  if (!hasVisibleWindows) {
    peregrineReopen();
  }
  return YES;
}

// Runs on the main queue once Wails has set up NSApp and its delegate.
void peregrineWatchApp(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    NSNotificationCenter *nc = [NSNotificationCenter defaultCenter];
    [nc addObserverForName:NSApplicationDidHideNotification object:nil queue:nil
                usingBlock:^(NSNotification *note) { peregrineAppHidden(1); }];
    [nc addObserverForName:NSApplicationDidUnhideNotification object:nil queue:nil
                usingBlock:^(NSNotification *note) { peregrineAppHidden(0); }];

    id delegate = [NSApp delegate];
    SEL sel = @selector(applicationShouldHandleReopen:hasVisibleWindows:);
    if (delegate == nil || [delegate respondsToSelector:sel]) {
      return; // nothing to add to, or the delegate handles it already
    }
    struct objc_method_description d =
        protocol_getMethodDescription(@protocol(NSApplicationDelegate), sel, NO, YES);
    if (d.types == NULL) {
      return;
    }
    if (class_addMethod([delegate class], sel, (IMP)peregrineShouldHandleReopen, d.types)) {
      // NSApplication looks at what its delegate implements when the delegate is set.
      [NSApp setDelegate:nil];
      [NSApp setDelegate:delegate];
    }
  });
}
