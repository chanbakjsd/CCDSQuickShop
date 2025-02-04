<script lang="ts">
	import { uploadImage } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import type { ChangeEventHandler } from 'svelte/elements';

	export let value: string;

	let fileSelect: HTMLInputElement;
	let resolveFileSelect: ((filelist: FileList) => void) | undefined = undefined;
	let rejectFileSelect: (() => void) | undefined = undefined;
	const upload = async () => {
		fileSelect.click();
		const files = await new Promise<FileList>((resolve, rej) => {
			resolveFileSelect = resolve;
			rejectFileSelect = rej;
		});
		if (files.length === 0) {
			return;
		}
		value = await uploadImage(files[0]);
	};

	const selectFile: ChangeEventHandler<HTMLInputElement> = (e) => {
		if (!e.currentTarget.files) {
			if (rejectFileSelect) rejectFileSelect();
			return;
		}
		if (resolveFileSelect) resolveFileSelect(e.currentTarget.files);
	};
</script>

<div class="flex gap-2">
	<input bind:value />
	<Button onClick={upload} size="md">Upload</Button>

	<input type="file" on:change={selectFile} bind:this={fileSelect} class="hidden" />
</div>

<style lang="postcss">
	input {
		@apply max-w-lg border border-black px-1;
	}
</style>
