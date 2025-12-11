import moment from "moment";
import { util } from "suimjs";

export default {
  findDimension(dimension = [], key) {
    if (!Array.isArray(dimension) && dimension.length == 0) return "";
    return dimension.find((e) => e.Key == key)?.Value;
  },
  isStatusDraft(status) {
    const _status = status ?? "";
    return ["DRAFT", ""].includes(_status);
  },
  isFormReadonly(status) {
    return !this.isStatusDraft(status);
  },
  isShowLog(status) {
    return !this.isStatusDraft(status);
  },
  isDisablePrintPreview(status) {
    return !["READY", "POSTED"].includes(status);
  },
  cloneObject(obj) {
    return JSON.parse(JSON.stringify(obj));
  },
  gridColumnConfig(obj) {
    const field = obj?.field ?? "field";
    const label = obj?.label ?? field;
    const placeHolder = obj?.input?.placeHolder ?? label;
    const kind = obj?.kind ?? "text";

    const cfg = {
      field: field,
      kind: kind,
      label: label,
      labelField: "",
      readType: "show",
      input: {
        lookupUrl: "",
        field: field,
        label: label,
        placeHolder: placeHolder,
        kind: kind,
      },
    };
    return { ...cfg, ...obj };
  },
  defaultPayloadBuilder(search, cfg, val, take = 20) {
    const _cfg = cfg ?? {
      lookupKey: "_id",
      lookupLabels: ["_id", "Name"],
      lookupSearchs: ["_id", "Name"],
      multiple: false,
    };

    let qp = {};
    qp.Take = take;
    qp.Sort = [_cfg.lookupLabels[0]];
    qp.Select = _cfg.lookupLabels;
    let idInSelect = false;
    const selectedFields = _cfg.lookupLabels.map((x) => {
      if (x == cfg.lookupKey) {
        idInSelect = true;
      }
      return x;
    });
    if (!idInSelect) {
      selectedFields.push(_cfg.lookupKey);
    }
    qp.Select = selectedFields;

    //setting search
    if (search.length > 0 && _cfg.lookupSearchs.length > 0) {
      if (_cfg.lookupSearchs.length == 1)
        qp.Where = {
          Field: _cfg.lookupSearchs[0],
          Op: "$contains",
          Value: [search],
        };
      else
        qp.Where = {
          Op: "$or",
          items: _cfg.lookupSearchs.map((el) => {
            return { Field: el, Op: "$contains", Value: [search] };
          }),
        };
    }

    if (!_cfg.multiple && val) {
      const whereExisting = { Op: "$eq", Field: _cfg.lookupKey, Value: val };
      if (qp.Where != undefined)
        qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
      else qp.Where = { Op: "$or", items: [whereExisting] };
    } else if (_cfg.multiple && val && val.length > 0) {
      const whereExisting =
        _cfg.modelValue.length == 1
          ? { Op: "$eq", Field: _cfg.lookupKey, Value: val[0] }
          : {
              Op: "$or",
              items: val.map((el) => {
                return { Field: _cfg.lookupKey, Op: "$eq", Value: el };
              }),
            };

      if (val.length > take) {
        qp.Take = props.modelValue.length + 1;
      }

      if (qp.Where != undefined) {
        qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
      } else {
        qp.Where = { Op: "$or", items: [whereExisting] };
      }
      // qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
    }
    return qp;
  },
  payloadBuilderDimension(list, val, mutliple = false, search) {
    const cfg = {
      lookupKey: "_id",
      lookupLabels: ["Label"],
      lookupSearchs: ["_id", "Label"],
      mutliple: mutliple,
    };
    let qp = this.defaultPayloadBuilder(search, cfg, val);

    if (qp.Where != undefined) {
      const items = [{ Op: "$contains", Field: `_id`, Value: list }];

      items.push(qp.Where);
      qp.Where = {
        Op: "$and",
        items: items,
      };
    } else {
      qp.Where = { Op: "$contains", Field: `_id`, Value: list };
    }
    return qp;
  },
  payloadBuilderTaxCodes(search, cfg, val, modules) {
    let qp = this.defaultPayloadBuilder(search, cfg, val);
    if (qp.Where != undefined) {
      const items = [{ Op: "$contains", Field: `Modules`, Value: [modules] }];
      items.push(qp.Where);
      qp.Where = {
        Op: "$and",
        items: items,
      };
    } else {
      qp.Where = { Op: "$contains", Field: `Modules`, Value: [modules] };
    }
    return qp;
  },
  payloadBuilderSpareAsset(search, cfg, val, groupid) {
    let qp = this.defaultPayloadBuilder(search, cfg, val);
    if (qp.Where != undefined) {
      const items = [{ Op: "$nin", Field: `GroupID`, Value: groupid }];
      items.push(qp.Where);
      qp.Where = {
        Op: "$and",
        items: items,
      };
    } else {
      qp.Where = { Op: "$nin", Field: `GroupID`, Value: groupid };
    }
    return qp;
  },
  ItemVarian(ItemID, SKU) {
    let ItemVarian = "";
    if (ItemID && SKU) {
      ItemVarian = `${ItemID}~~${SKU}`;
    } else if (ItemID) {
      ItemVarian = ItemID;
    }
    return ItemVarian;
  },
  genFilterDimension(dimension) {
    const g = Object.groupBy(dimension, (obj) => obj.Key);
    const filters = [];
    Object.keys(g).forEach((key) => {
      const val = g[key][0].Value;
      if (Array.isArray(val) && val.length > 0) {
        filters.push({
          Op: "$and",
          Items: [
            {
              Op: "$eq",
              Field: "Dimension.Key",
              Value: key,
            },
            {
              Op: "$in",
              Field: "Dimension.Value",
              Value: [...val],
            },
          ],
        });
      }
    });
    return filters;
  },
  formatFilterDate(date, endDay = false) {
    let dt = moment(date).set({
      hour: 0,
      minute: 0,
      second: 0,
      millisecond: 0,
    });
    if (endDay) {
      dt = moment(date).set({
        hour: 23,
        minute: 59,
        second: 59,
        millisecond: 999,
      });
    }
    return new Date(moment.utc(dt));
  },
  validateSiteEntryExpense(expense) {
    const r = expense.filter((o) => o.TotalAmount == 0);
    const valid = r.length == 0;
    if (!valid) {
      util.showError("Amount must be > 0");
    }
    return valid;
  },
  formatNumberWithDot(number) {
    const val = number ? number : 0;
    return new Intl.NumberFormat("id-ID", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(val);
  },
  updateTags(axios, payload) {
    axios.post(`/asset/update-tag`, payload).then((r) => {
      if (r.data != "success") {
        util.showError("error update tags");
      }
    });
  },

  updatePrint(axios, moduleid, Source, payload, cbOK, cbFalse) {
    axios.post(`${moduleid}/${Source}/update-print`, payload).then(
      (r) => {
        if (cbOK) {
          cbOK();
        }
      },
      (e) => {
        if (cbFalse) {
          cbFalse(emit);
        }
      }
    );
  },

  generateGridCfg(colum) {
    let addColm = [];
    for (let index = 0; index < colum.length; index++) {
      addColm.push({
        field: colum[index].field,
        kind: colum[index].kind,
        label: colum[index].label,
        readType: "show",
        labelField: "",
        width: colum[index].width,
        readOnly: colum[index].readOnly,
        input: {
          field: colum[index].field,
          label: colum[index].label,
          hint: "",
          hide: false,
          placeHolder: colum[index].label,
          kind: colum[index].kind,
          width: colum[index].width,
          readOnly: colum[index].readOnly,
        },
      });
    }
    return {
      fields: addColm,
      setting: {
        idField: "_id",
        keywordFields: ["_id", "Name"],
        sortable: ["_id"],
      },
    };
  },
  dateTimeNow(originalTime) {
    let originalDate = new Date(originalTime);
    let now = new Date();
    originalDate.setHours(now.getHours());
    originalDate.setMinutes(now.getMinutes());
    originalDate.setSeconds(now.getSeconds());
    return moment(originalDate).format();
  },
  getAssetUrl(id) {
    const url = 'https://bis-dev.kanosolution.app/v1/'//import.meta.env.VITE_API_URL;
    return url + "asset/view?id=" + id;
  },
  convertToWordsIDR(angka) {
    const satuan = ["", "Satu", "Dua", "Tiga", "Empat", "Lima", "Enam", "Tujuh", "Delapan", "Sembilan"];
    const belasan = ["Sepuluh", "Sebelas", "Dua Belas", "Tiga Belas", "Empat Belas", "Lima Belas", "Enam Belas", "Tujuh Belas", "Delapan Belas", "Sembilan Belas"];
    const puluhan = ["", "", "Dua Puluh", "Tiga Puluh", "Empat Puluh", "Lima Puluh", "Enam Puluh", "Tujuh Puluh", "Delapan Puluh", "Sembilan Puluh"];
    const ribuan = ["", "Ribu", "Juta", "Miliar", "Triliun"];

    if (isNaN(angka) || angka < 0) {
        return "";
    }

    if (angka === 0) {
        return "Nol Rupiah";
    }

    function terbilang(angka) {
      let hasil = "";
      if (angka < 10) {
          hasil = satuan[angka];
      } else if (angka < 20) {
          hasil = belasan[angka - 10];
      } else if (angka < 100) {
          hasil = puluhan[Math.floor(angka / 10)] + (angka % 10 !== 0 ? " " + satuan[angka % 10] : "");
      } else if (angka < 1000) {
          hasil = (angka < 200 ? "Seratus" : satuan[Math.floor(angka / 100)] + " Ratus") +
              (angka % 100 !== 0 ? " " + terbilang(angka % 100) : "");
      } else {
        for (let i = ribuan.length - 1; i >= 0; i--) {
          const pembagi = Math.pow(1000, i);
          if (angka >= pembagi) {
              const banyak = Math.floor(angka / pembagi);
              hasil += (banyak === 1 && i === 1 ? "Seribu" : terbilang(banyak) + " " + ribuan[i]) +
                  (angka % pembagi !== 0 ? " " + terbilang(angka % pembagi) : "");
              break;
          }
        }
      }
      return hasil.trim();
  }

    // Pisahkan angka sebelum dan sesudah koma
    const [angkaUtama, angkaKoma] = angka.toString().split(".");
    let hasil = terbilang(parseInt(angkaUtama)) + " Rupiah";

    if (angkaKoma) {
        const komaTerbilang = angkaKoma.split("").map(digit => satuan[parseInt(digit)]).join(" ");
        hasil += " dan " + komaTerbilang + " Sen";
    }

    return hasil;
  }
};
