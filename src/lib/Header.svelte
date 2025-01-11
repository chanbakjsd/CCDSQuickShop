<script lang="ts">
	export let admin = false;
</script>

<div
	class="sticky top-0 -my-2 flex items-center justify-center gap-2 bg-white pb-2 md:justify-start"
>
	<img src="https://ntuscds.com/scse-logo/scds-logo.png" class="size-16" alt="Logo for SCDS Club" />
	<p class="text-2xl font-bold text-brand">
		Merch Store
		{#if admin}(Admin){/if}
	</p>
</div>
