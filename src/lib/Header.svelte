<script lang="ts">
	export let admin = false;
	export let cls = '';
</script>

<div
	class={`${cls} sticky left-0 top-0 -mb-1 -mt-2 flex w-screen items-center justify-center gap-2 border-b border-gray-400 px-4 pb-1 md:w-full md:justify-start`}
>
	<img src="https://ntuscds.com/scse-logo/scds-logo.png" class="size-16" alt="Logo for SCDS Club" />
	<p class="text-2xl font-bold text-brand">
		Merch Store
		{#if admin}(Admin){/if}
	</p>
</div>
