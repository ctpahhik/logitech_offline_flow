// hook.c
#include <windows.h>
#include "hook.h"

extern void onMouseMove(int x, int y);

static HHOOK hook;
static MSG msg;

static LRESULT CALLBACK MouseHookProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0 && wParam == WM_MOUSEMOVE) {
        MSLLHOOKSTRUCT *pMouse = (MSLLHOOKSTRUCT *)lParam;
        onMouseMove((int)pMouse->pt.x, (int)pMouse->pt.y);
    }
    return CallNextHookEx(hook, nCode, wParam, lParam);
}

static void SetMouseHook() {
    hook = SetWindowsHookEx(WH_MOUSE_LL, MouseHookProc, NULL, 0);
    if (hook == NULL) {
        MessageBox(NULL, "Failed to install hook!", "Error", MB_ICONERROR);
        return;
    }

    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}