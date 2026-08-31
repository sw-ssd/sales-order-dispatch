export default function ForbiddenPage() {
  return (
    <main class="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div class="w-full max-w-sm rounded-lg bg-white p-6 text-center shadow">
        <p class="text-5xl font-bold text-gray-300">403</p>
        <h1 class="mt-2 text-lg font-semibold text-gray-900">無權限存取</h1>
        <p class="mt-1 text-sm text-gray-500">你的角色不允許存取此頁面</p>
        <a
          href="/"
          class="mt-4 inline-block rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          回首頁
        </a>
      </div>
    </main>
  );
}
