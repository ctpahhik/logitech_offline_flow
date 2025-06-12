#include <windows.h>
#include "../system_api.h"

extern void OnMouseMove(int x, int y);
extern void OnEnabledHotKey();

static HHOOK hook;
static MSG msg;

static LRESULT CALLBACK MouseHookProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0 && wParam == WM_MOUSEMOVE) {
        MSLLHOOKSTRUCT *pMouse = (MSLLHOOKSTRUCT *)lParam;
        OnMouseMove((int)pMouse->pt.x, (int)pMouse->pt.y);
    }
    return CallNextHookEx(hook, nCode, wParam, lParam);
}

static int SetMouseHook() {
    hook = SetWindowsHookEx(WH_MOUSE_LL, MouseHookProc, NULL, 0);
    if (hook == NULL) {
        return -1;//"Failed to install hook!";
    }

    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
    return 0;
}

/*static void RegisterEnabledHotKey() {
}*/

/*static void UnregisterEnabledHotKey() {
}*/

/*
static void MoveMouse() {
}*/
