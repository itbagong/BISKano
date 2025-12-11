import { defineStore } from "pinia"
import { util } from "suimjs"

export const notifStore = defineStore("notif", {
    state: () => {
        return {
            messages: []
        }
    },

    actions: {
        add(obj) {
            const msgid = util.uuid()
            var message = obj.message
            if (obj.kind=='error') message = util.getErrorMessage(obj.message)
            this.messages.push({msgid: msgid, kind: obj.kind, message: message, dispatchTime: new Date()})

            setTimeout(() => {
                this.remove(msgid)
            }, 3000)
        },

        remove (id) {
            this.messages = this.messages.filter(el => el.msgid!=id)
        }
    }
})

