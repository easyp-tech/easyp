<script setup lang="ts">

import {ref} from 'vue'

import {GetLatestRelease} from "../../src/version/version";

const Version = ref(GetLatestRelease());

</script>