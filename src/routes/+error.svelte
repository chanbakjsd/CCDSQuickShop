<script lang="ts">
	import { page } from '$app/state'
</script>

<div class="flex h-screen w-screen flex-col items-center justify-center gap-2">
	<img src="https://ntuscds.com/scds-logo-blue.png" class="mb-2 size-32" alt="Logo for SCDS Club" />
	<h1 class="text-xl">Error</h1>
	<p class="text-center text-sm">
		{page.error?.message || 'Additional error details are unavailable.'}<br />
		If you did not expect this error and the error persists, please contact the SCDS Club with details
		on how to reproduce the issue.
	</p>
	<a href="/" class="text-blue-700 underline">Return to Home</a>
</div>
