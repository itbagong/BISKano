<template>
    <template v-if="kind == 'skeleton'">
      <div
        v-if="skeletonKind == 'input'"
        class="skeleton-loading skeleton-loading__input"
      ></div>
      <div
        v-else-if="skeletonKind == 'text'"
        class="skeleton-loading skeleton-loading__text"
      ></div>
      <div
        v-else-if="skeletonKind == 'textarea'"
        class="skeleton-loading skeleton-loading__textarea"
      ></div>
      <div
        v-else-if="skeletonKind == 'card'"
        class="skeleton-loading skeleton-loading__card"
      ></div>
      <div
        v-else-if="skeletonKind == 'list'"
        class="w-full skeleton-loading__list mb-1"
      >
        <div class="skeleton-loading skeleton-loading__list--header"></div>
        <div class="skeleton-loading skeleton-loading__list--sub-header"></div>
        <div class="skeleton-loading skeleton-loading__list--content"></div>
      </div>
      <div v-else class="skeleton-loading skeleton-loading__default"></div>
    </template>
    <div
      v-else-if="kind == 'circle'"
      class="circle-loading"
      :class="{
        'size-md': circleSize == 'md',
        'size-sm': circleSize == 'sm',
      }"
    >
      <span></span>
    </div>
    <div v-else-if="kind == 'linier'" class="linear-loading">
      <div class="indeterminate"></div>
    </div>
    <div v-else class="loader-wrapper">
        <div class="bus-wrapper">
          <div class="bus">
            <div class="bus-container"></div>
            <div class="glases"></div>
            <div class="bonet"></div>
            <div class="base"></div>
            <div class="base-aux"></div>
            <div class="wheel-back"></div>
            <div class="wheel-front"></div>
            <div class="smoke"></div>
          </div>
        </div>
    </div>
</template>
<script setup>
const props = defineProps({
  kind: { type: String, default: "linier" }, //skeleton|circle|linier
  skeletonKind: { type: String, default: "" }, //input|textarea|list
  circleSize: { type: String, default: "md" },
});
</script>
<style scoped>
.skeleton-loading {
  @apply mb-2;
  width: 100%;
  background-color: #ddd;
  border-radius: 2px;
  animation: pulse-bg 1s infinite;
}
.skeleton-loading__default {
  min-height: 100px;
  height: 100%;
}
.skeleton-loading__text {
  height: 28px;
  margin-bottom: 1px !important;
}
.skeleton-loading__input {
  height: 32px;
}
.skeleton-loading__textarea {
  height: 100px;
}
.skeleton-loading__list {
  @apply flex flex-col gap-[1px];
}
.skeleton-loading__list--header {
  height: 14px;
}
.skeleton-loading__list--sub-header {
  width: 50%;
  height: 14px;
}
.skeleton-loading__list--content {
  width: 75%;
  height: 14px;
}

@keyframes pulse-bg {
  0% {
    background-color: #ddd;
  }
  50% {
    background-color: #d0d0d0;
  }
  100% {
    background-color: #ddd;
  }
}

.linear-loading {
  @apply overflow-hidden w-full h-[4px] bg-primary;
}
.indeterminate {
  position: relative;
  width: 100%;
  height: 100%;
}
.indeterminate:before {
  content: "";
  position: absolute;
  height: 100%;
  background-color: hsl(166, 20%, 62%);
  animation: indeterminate_first 1.5s infinite ease-out;
}
.indeterminate:after {
  @apply bg-secondary;
  content: "";
  position: absolute;
  height: 100%;
  animation: indeterminate_second 1.5s infinite ease-in;
}
@keyframes indeterminate_first {
  0% {
    left: -100%;
    width: 100%;
  }
  100% {
    left: 100%;
    width: 10%;
  }
}

@keyframes indeterminate_second {
  0% {
    left: -150%;
    width: 100%;
  }
  100% {
    left: 100%;
    width: 10%;
  }
}

/* Loader 2 */
.circle-loading {
  display: block;

  -webkit-animation: circle-loading-2-1 3s linear infinite;
  animation: circle-loading-2-1 3s linear infinite;
}
@-webkit-keyframes circle-loading-2-1 {
  0% {
    -webkit-transform: rotate(0deg);
  }
  100% {
    -webkit-transform: rotate(360deg);
  }
}
@keyframes circle-loading-2-1 {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
.circle-loading span {
  display: block;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
  margin: auto;
  -webkit-animation: circle-loading-2-2 1.5s cubic-bezier(0.77, 0, 0.175, 1)
    infinite;
  animation: circle-loading-2-2 1.5s cubic-bezier(0.77, 0, 0.175, 1) infinite;
}
@-webkit-keyframes circle-loading-2-2 {
  0% {
    -webkit-transform: rotate(0deg);
  }
  100% {
    -webkit-transform: rotate(360deg);
  }
}
@keyframes circle-loading-2-2 {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
.circle-loading span::before {
  content: "";
  display: block;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
  margin: auto;
  border: 3px solid transparent;
  border-top: 3px solid #fff;
  border-radius: 50%;
  -webkit-animation: circle-loading-2-3 1.5s cubic-bezier(0.77, 0, 0.175, 1)
    infinite;
  animation: circle-loading-2-3 1.5s cubic-bezier(0.77, 0, 0.175, 1) infinite;
}
@-webkit-keyframes circle-loading-2-3 {
  0% {
    -webkit-transform: rotate(0deg);
  }
  100% {
    -webkit-transform: rotate(360deg);
  }
}
@keyframes circle-loading-2-3 {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.circle-loading span::after {
  @apply border-[3px] border-primary;
  content: "";
  display: block;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
  margin: auto;
  border-radius: 50%;
}

.circle-loading.size-md {
  height: 32px;
  width: 32px;
}
.circle-loading.size-md span {
  clip: rect(16px, 32px, 32px, 0);
}
.circle-loading.size-md span,
.circle-loading.size-md span::after,
.circle-loading.size-md span::before {
  height: 32px;
  width: 32px;
}

.circle-loading.size-md {
  height: 22px;
  width: 22px;
}
.circle-loading.size-sm span {
  clip: rect(12px, 22px, 22px, 0);
}
.circle-loading.size-sm span,
.circle-loading.size-sm span::after,
.circle-loading.size-sm span::before {
  height: 22px;
  width: 22px;
}
.loader-wrapper{
  position:fixed;
  z-index:1090;
  height:100vh;
  width:100vw;
  background-color:rgba(240,240,240,0.5) !important;
}

.bus-wrapper{
  height: 200px;
  width: 200px;
  /* border: 5px solid #4CAF50; */
  position:absolute;
  top:30%;
  left:50%;
  transform:translateX(-50%) translateY(-50%) scale(0.8);
   background:transparent;
  animation:bg 0.5s linear infinite;
  /* border-bottom:3px solid #404143; */
  /* border-radius: 100%; */
  overflow:hidden;
/*   box-shadow:inset 0px 0px 10px 4px rgba(0,0,0,0.3),inset 0px 0px 5px 0px #4CAF50; */
}

.bus-wrapper:after{
    content:'Loading...';
    font-size:20px;
    position:absolute;
    bottom:0px;
    text-align:center;
    width:100%;
    border-top:1px solid  var(--color-primary);
    background: var(--color-primary));
    background: -moz-linear-gradient(left, var(--color-primary) 0%, var(--color-secondary) 100%);
    background: -webkit-linear-gradient(left, var(--color-primary) 0%, var(--color-secondary) 100%);
    background: linear-gradient(to right, var(--color-primary) 0%, var(--color-secondary) 100%);
    filter: progid:DXImageTransform.Microsoft.gradient( startColorstr='#1efcc8', endColorstr='#1dd3d6',GradientType=1 );
    color:white;
    padding-top:4px;
    padding-bottom:8px;
    animation: bg 3s linear infinite;
}

.bus{
  height:110px;
  width:150px;
  position:absolute;
  bottom:48px;
  left: calc(50% + 17px);
  transform: translateX(-50%);
  
}

 
.bus > .glases{
    background: rgb(40,181,245);
    background: -moz-linear-gradient(-45deg, rgba(40,181,245,1) 0%, rgba(40,181,245,1) 50%, rgba(2,153,227,1) 52%, rgba(2,153,227,1) 100%);
    background: -webkit-linear-gradient(-45deg, rgba(40,181,245,1) 0%,rgba(40,181,245,1) 50%,rgba(2,153,227,1) 52%,rgba(2,153,227,1) 100%);
    background: linear-gradient(135deg, #514f4f 0%,#514f4f 50%,#333131 52%,#333131 100%);
    filter: progid:DXImageTransform.Microsoft.gradient( startColorstr='#28b5f5', endColorstr='#0299e3',GradientType=1 );
    position:absolute;
    height:25px;
    width:143.9px;
    border:4px solid var(--color-primary);
    border-bottom:none;
    top:35.5px;
    left:-19px;
    border-top-right-radius:6px;
    animation: updown-half 0.4s linear infinite;
}
.bus > .glases:after{
    content:'';
    display:block;
    background-color: var(--color-primary);
    height: 21px;
    width: 4px;
    position:absolute;
    right:-6px;
    bottom:0px;
    border-radius:10px / 15px;
    border-bottom-right-radius:0px;
    border-bottom-left-radius:0px;
    border-top-left-radius:0px;
  
}

.bus > .glases:before{
    content:'';
    display:block;
    background-color: var(--color-primary);
    height: 21px;
    width: 4px;
    position:absolute;
    left:102px;
    bottom:0px;
    /*   border-top-right-radius:4px; */
}

.bus > .bonet{
    background-color:var(--color-primary);
    position:absolute;
    width:153.8px;
    height:15px;
    top:64px;
    left:-19px;
    z-index:-1;
    animation: updown 0.4s linear infinite;
}

.bus > .bonet:after{
    content:'';
    display:block;
    background: rgb(255,255,255);
    background: -moz-linear-gradient(-45deg, rgba(255,255,255,1) 0%, rgba(241,241,241,1) 50%, rgba(225,225,225,1) 51%, rgba(246,246,246,1) 100%);
    background: -webkit-linear-gradient(-45deg, rgba(255,255,255,1) 0%,rgba(241,241,241,1) 50%,rgba(225,225,225,1) 51%,rgba(246,246,246,1) 100%);
    background: linear-gradient(135deg, rgba(255,255,255,1) 0%,rgba(241,241,241,1) 50%,rgba(225,225,225,1) 51%,rgba(246,246,246,1) 100%);
    filter: progid:DXImageTransform.Microsoft.gradient( startColorstr='#ffffff', endColorstr='#f6f6f6',GradientType=1 );
    height:10px;
    width:6px;
    position:absolute;
    right:0px;
    bottom:2px;
    border-top-left-radius:4px;
  
}

.bus > .base{
    position:absolute;
    background-color:#595759;
    width:134px;
    height:15px;
    border-top-right-radius:10px;
    top:70px;
    left:-15px;
    animation: updown 0.4s linear infinite;
}

.bus > .base:before{
    content:'';
    display:block;
    background-color:#E54A18;
    height:20px;
    width:5px;
    position:absolute;
    left:-4px;
    bottom:0px;
    border-bottom-left-radius:4px;
}

.bus > .base:after{
    content:'';
    display:block;
    background-color:#595759;
    height:10px;
    width:20px;
    position:absolute;
    right:-16px;
    bottom:0px;
    border-bottom-right-radius:4px;
    z-index:-1;
}

.bus > .base-aux{
    width:3px;
    height:26px;
    background-color:var(--color-primary);
    position:absolute;
    top:38px;
    left:25px;
    /*   border-bottom-right-radius:4px; */
    animation: updown-half 0.4s linear infinite;
}
.bus > .wheel-back{
    left:20px
}

.bus > .wheel-front{
    left:95px;
}

.bus > .wheel-back,.bus > .wheel-front{
    border-radius:100%;
    position:absolute;
    background: rgb(84,110,122);
    background: -moz-linear-gradient(-45deg, rgba(84,110,122,1) 0%, rgba(84,110,122,1) 49%, rgba(68,90,100,1) 52%, rgba(68,90,100,1) 100%);
    background: -webkit-linear-gradient(-45deg, rgba(84,110,122,1) 0%,rgba(84,110,122,1) 49%,rgba(68,90,100,1) 52%,rgba(68,90,100,1) 100%);
    background: linear-gradient(135deg, rgba(84,110,122,1) 0%,rgba(84,110,122,1) 49%,rgba(68,90,100,1) 52%,rgba(68,90,100,1) 100%);
    filter: progid:DXImageTransform.Microsoft.gradient( startColorstr='#546e7a', endColorstr='#445a64',GradientType=1 );
    top:75px;
    height:22px;
    width:22px;
    animation:spin 0.6s linear infinite;
}

.bus > .wheel-back:before,.bus > .wheel-front:before{
    content:'';
    border-radius:100%;
        left:5px;
    top:5px;
    position:absolute;
        background: rgb(175,189,195);
    background: -moz-linear-gradient(-45deg, rgba(175,189,195,1) 0%, rgba(175,189,195,1) 50%, rgba(143,163,173,1) 51%, rgba(143,163,173,1) 100%);
    background: -webkit-linear-gradient(-45deg, rgba(175,189,195,1) 0%,rgba(175,189,195,1) 50%,rgba(143,163,173,1) 51%,rgba(143,163,173,1) 100%);
    background: linear-gradient(135deg, rgba(175,189,195,1) 0%,rgba(175,189,195,1) 50%,rgba(143,163,173,1) 51%,rgba(143,163,173,1) 100%);
    filter: progid:DXImageTransform.Microsoft.gradient( startColorstr='#afbdc3', endColorstr='#8fa3ad',GradientType=1 );
    height:12px;
    width:12px; 
}

@keyframes spin {
  50%{
    top:76px;
  }
    100% {
        transform: rotate(360deg);
    }
}

@keyframes container {
  
  30%{
    transform:rotate(1deg);
  }
  50%{
    top:11px;
  }
  
  70%{
    top:10px;
    transform:rotate(-1deg);
  }
}

.bus > .smoke{
  position:absolute;
  background-color:#AFBDC3;
  border-radius:100%;
  width:8px;
  height:8px;
  top:90px;
  left:6px;
  animation: fade 0.4s linear infinite;
  opacity:0;
}

.bus > .smoke:after{
  content:'';
  position:absolute;
  background-color:RGB(143,163,173);
  border-radius:100%;
  width:6px;
  height:6px;
  top:-4px;
  left:4px;
}

.bus > .smoke:before{
  content:'';
  position:absolute;
  background-color:RGB(143,163,173);
  border-radius:100%;
  width:4px;
  height:4px;
  top:-2px;
  left:0px;
}

@keyframes fade {
  
  30%{
    opacity:0.3;
    left:7px;
  }
  50%{
    opacity:0.5;
    left:6px;
  }
  
  70%{
   opacity:0.1;
    left:4px;
  }
  
  90%{
    opacity:0.4;
    left:2px;
  }
}

@keyframes bg {
  from{
    background-position-x:0px;
  }
  to{
    background-position-x:-400px;
  }
}

@keyframes updown {
  50%{
    transform:translateY(-20%);
  }
  
  70%{
    transform:translateY(-10%);
  }
}

@keyframes updown-half{
    50%{
    transform:translateY(-10%);
  }
  
  70%{
    transform:translateY(-5%);
  }
}


</style>