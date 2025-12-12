/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  PermissionsAndroid,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import { useActions } from '@overmind/index';
import container from 'components/container';
import s from '@components/styles/index';
import { Colors, Mixins, Typography } from 'utils';
import Card from 'components/card';
import { Divider, Icon } from '@ui-kitten/components';
import RNFetchBlob from 'rn-fetch-blob';
import { API_URL } from '@env';
import { ALERT_TYPE, Toast } from 'react-native-alert-notification';
import moment from 'moment';
import { default as theme } from '../../custom-theme.json';

import ModalApproval from '../approval/modal-approval';
import ModalAttachment from '../approval-detail/modal-attachment';
import { ClipboardClose, ClipboardTick } from 'iconsax-react-native';

type Props = {
  navigation: any;
  route: any;
};
const getPrefix = (type: string) => {
  switch (type) {
    case "Purchase Request":
      return "PR";
    case "Purchase Order":
      return "PO";
    case "Movement In":
      return "MI";
    case "Movement Out":
      return "MO";
    case "Item Transfer":
      return "IT";
    case "Item Request":
      return "IR";
    case "Work Request":
      return "WR";
    case "Work Order":
      return "WO";
    case "Good Receive":
      return "GR";
    case "Good Issuance":
      return "GI";
    default:
      return type;
  }
}
const Layout = (props: Props) => {
  const { route, navigation } = props;
  const { findPreview, getsAssetByJournal, postApproval } = useActions();
  const [data, setData] = React.useState({} as any);
  const [loading, setLoading] = React.useState(false);

  const [showButton, setShowButton] = React.useState(route.params.Status === 'PENDING');
  const [showApproval, setShowApproval] = React.useState(false);
  const [operation, setOperation] = React.useState('');
  const [showAttachment, setShowAttachment] = React.useState(false);
  const [attachments, setAttachment] = React.useState([]);

  const onShowAttachment = (item: any) => {
    getsAssetByJournal({
      JournalType: item.SourceType,
      JournalID: item.SourceID,
      Tags: [`${getPrefix(item.SourceType)}_${item.SourceID}`],
    })
      .then((res: any) => {
        setAttachment(res);
        setShowAttachment(true);
      })
      .catch(e => {
        return Toast.show({
          type: ALERT_TYPE.DANGER,
          title: 'Error',
          textBody: e,
        });
      });
  };
  const print = async () => {
    const { config, fs } = RNFetchBlob;
    const downloadPath =
      fs.dirs.DownloadDir +
      `/${route.params.SourceType}_${moment().format('DDMMYYYYhhmmss')}.pdf`;
    const url = `${API_URL}/fico/postingprofile/preview-download-as-pdf?SourceType=${route.params.SourceType}&SourceJournalID=${route.params.SourceJournalID}`;
    // if (!__DEV__) {
    //   const granted = await PermissionsAndroid.request(
    //     PermissionsAndroid.PERMISSIONS.WRITE_EXTERNAL_STORAGE,
    //     {
    //       title: 'Storage Permission Required',
    //       message: 'App needs access to your storage to download the file',
    //       buttonPositive: 'Ok',
    //     },
    //   );
    //   if (granted !== PermissionsAndroid.RESULTS.GRANTED) {
    //     Toast.show({
    //       type: ALERT_TYPE.DANGER,
    //       title: 'Permission Denied!',
    //       textBody: 'You need to give storage permission to download the file',
    //     });
    //     return;
    //   }
    // }
    config({
      fileCache: true,
      addAndroidDownloads: {
        useDownloadManager: true,
        notification: true,
        path: downloadPath,
        description: 'Downloading PDF',
      },
    })
      .fetch('GET', url)
      .then(() => {
        Toast.show({
          type: ALERT_TYPE.SUCCESS,
          title: 'Download Success',
          textBody: 'PDF downloaded successfully',
        });
      })
      .catch((error: any) => {
        Toast.show({
          type: ALERT_TYPE.DANGER,
          title: 'Download Failed!',
          textBody: 'PDF download failed: ' + error.message,
        });
      });
  };
  React.useLayoutEffect(() => {
    navigation.setOptions({
      headerRight: () => (
        <View style={{ flexDirection: 'row', gap: 5 }}>
          <TouchableOpacity style={styles.containerPrint} onPress={() => onShowAttachment(route.params.Item)}>
            <Icon
              style={{
                width: Mixins.scaleSize(20),
                height: Mixins.scaleSize(20),
              }}
              fill={Colors.PRIMARY.green}
              name="attach-outline"
            />
          </TouchableOpacity>
          <TouchableOpacity style={styles.containerPrint} onPress={() => print()}>
            <Icon
              style={{
                width: Mixins.scaleSize(20),
                height: Mixins.scaleSize(20),
              }}
              fill={Colors.PRIMARY.green}
              name="printer-outline"
            />
          </TouchableOpacity>
        </View>
      ),
    });
  }, [navigation]);
  const init = () => {
    setLoading(true);
    findPreview(route.params).then(res => {
      setData(res);
    }).finally(() => {
      setLoading(false);
    })
  }
  React.useEffect(() => {
    init();
    return () => { };
  }, []);

  const pairsComp = (items: any) => {
    const pairs = [];
    for (let i = 0; i < items.length; i += 2) {
      if (items.slice(i, i + 2).find((o: any) => o !== '')) {
        pairs.push(items.slice(i, i + 2));
      }
    }
    return pairs;
  };
  const mappingData = (items: any[]) => {
    let keys = items[0];
    let result = [];
    for (let i = 1; i < items.length; i++) {
      let obj: any = {};
      for (let j = 0; j < items[i].length; j++) {
        obj[keys[j]] = items[i][j];
      }
      result.push(obj);
    }

    return result;
  };
  const formatHeader = (v: string) => {
    if (v) {
      let splited = v?.split(':');
      return splited.length > 1 ? splited[0] : v;
    }
    return v;
  };

  const rendderItem = (each: any) => {
    const arr: any[] = [];
    Object.entries(each).forEach(([key, value]: any) => {
      arr.push(
        <View key={key} style={s.row}>
          <Text
            style={{
              flex: 1,
              ...Typography.textMdPlus,
              color: Colors.BLACK,
            }}>
            {formatHeader(key)}
          </Text>
          <Text
            style={{
              flex: 1.5,
              ...Typography.textMdPlus,
              color: Colors.BLACK,
            }}>
            : {value}
          </Text>
        </View>,
      );
    });
    return arr;
  };
  const onSubmitPost = (text: string) => {
    setShowApproval(false);
    setLoading(true);
    const filteredItem = [route.params.Item];
    let payloadFico: any[] = [];
    let payloadSCM: any[] = [];
    let payloadMFG: any[] = [];
    let post = [];

    filteredItem.forEach((item: any) => {
      if (
        [
          'INVENTORY',
          'Inventory Receive',
          'Inventory Issuance',
          'Transfer',
          'Item Request',
          'Purchase Order',
          'Purchase Request',
          'Asset Acquisition',
        ].includes(item.SourceType)
      ) {
        payloadSCM.push({
          JournalID: item.SourceID,
          JournalType: item.SourceType,
          Op: operation,
          Text: text,
        });
      } else if (
        [
          'Work Request',
          'Work Order',
          'Work Order Report Consumption',
          'Work Order Report Resource',
          'Work Order Report Output',
        ].includes(item.SourceType)
      ) {
        payloadMFG.push({
          JournalID: item.SourceID,
          JournalType: item.SourceType,
          Op: operation,
          Text: text,
        });
      } else {
        payloadFico.push({
          JournalID: item.SourceID,
          Op: operation,
          Text: text,
        });
      }
    });
    if (payloadFico.length > 0) {
      post.push(postApproval({ type: 'fico', payload: payloadFico }));
    }
    if (payloadSCM.length > 0) {
      post.push(postApproval({ type: 'scm', payload: payloadSCM }));
    }
    if (payloadMFG.length > 0) {
      post.push(postApproval({ type: 'mfg', payload: payloadMFG }));
    }
    Promise.all(post)
      .then(() => {
        Toast.show({
          type: ALERT_TYPE.SUCCESS,
          title: 'Success',
          textBody: `Data has been ${operation === 'Approve' ? 'Approved' : 'Rejected'
            }`,
        });
        init();
        setShowButton(false);
      })
      .finally(() => {
        setLoading(false);
      });
  };
  const onOpenApproval = (op: string) => {
    setOperation(op);
    setTimeout(() => {
      setShowApproval(true);
    }, 100);
  };
  return (
    <View style={s.container}>
      <View style={{ flex: 1 }}>
        <ScrollView showsVerticalScrollIndicator={false}>
          <View style={styles.section}>
            {data?.HeaderMobile?.Data !== null ? (
              <>
                {data?.HeaderMobile?.Data?.map((header: any, i: number) => (
                  <View key={i} style={styles.header}>
                    {pairsComp(header).map((item: any, index: number) => {
                      return (
                        <View key={index} style={[styles.headerItemContainer]}>
                          <Text style={styles.headerText}>{item[0]}</Text>
                          <Text style={styles.headerText}>{item[1]}</Text>
                        </View>
                      );
                    })}
                  </View>
                ))}
              </>
            ) : (
              <>
                {data?.Header?.Data?.map((header: any, i: number) => (
                  <View key={i} style={styles.header}>
                    {pairsComp(header).map((item: any, index: number) => {
                      return (
                        <View key={index} style={[styles.headerItemContainer]}>
                          <Text style={styles.headerText}>{item[0]}</Text>
                          <Text style={styles.headerText}>{item[1]}</Text>
                        </View>
                      );
                    })}
                  </View>
                ))}
              </>
            )}
          </View>
          {data?.Sections?.map((item: any, i: number) => (
            <Card key={i} containerStyle={styles.card}>
              <View style={styles.cardHeader}>
                <Text style={styles.cardHeaderTitle}>{item.Title}</Text>
              </View>
              <View style={styles.cardBody}>
                {mappingData(item.Items).map((each: any, ii: number) => {
                  return (
                    <View
                      key={ii}
                      style={{
                        marginBottom: Mixins.scaleSize(10),
                        paddingBottom: Mixins.scaleSize(5),
                      }}>
                      {rendderItem(each).map(o => {
                        return o;
                      })}
                      <Divider style={{ marginTop: Mixins.scaleSize(10) }} />
                    </View>
                  );
                })}
              </View>
            </Card>
          ))}

          <View style={styles.section}>
            {data?.HeaderMobile?.Footer !== null ? (
              <>
                {data?.HeaderMobile?.Footer?.map((footer: any, i: number) => (
                  <View key={i} style={styles.header}>
                    {pairsComp(footer).map((item: any, index: number) => {
                      return (
                        <View key={index} style={[styles.headerItemContainer]}>
                          <Text style={styles.headerText}>{item[0].trim()}</Text>
                          <Text style={styles.headerText}>
                            {formatHeader(item[1])}
                          </Text>
                        </View>
                      );
                    })}
                  </View>
                ))}
              </>
            ) : (
              <>
                {data?.Header?.Footer?.map((footer: any, i: number) => (
                  <View key={i} style={styles.header}>
                    {pairsComp(footer).map((item: any, index: number) => {
                      return (
                        <View key={index} style={[styles.headerItemContainer]}>
                          <Text style={styles.headerText}>{item[0]}</Text>
                          <Text style={styles.headerText}>
                            {formatHeader(item[1])}
                          </Text>
                        </View>
                      );
                    })}
                  </View>
                ))}
              </>
            )}
          </View>
        </ScrollView>
      </View>
      {showButton && (
        <View
          style={{
            ...s.row,
            gap: Mixins.scaleSize(10),
            // paddingVertical: Mixins.scaleSize(10),
          }}>
          <TouchableOpacity
            onPress={() => {
              onOpenApproval('Approve');
            }}
            disabled={loading}
            style={{
              ...styles.buttonAction,
              backgroundColor: theme['color-primary-500'],
            }}>
            <ClipboardTick size="32" color="white" />
            <Text
              style={{
                ...styles.labelAction,
                color: Colors.WHITE,
              }}>
              Approve
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            onPress={() => {
              onOpenApproval('Reject');
            }}
            disabled={loading}
            style={{
              ...styles.buttonAction,
              backgroundColor: Colors.SHADES.gray[500],
            }}>
            <ClipboardClose size="32" color="white" />
            <Text
              style={{
                ...styles.labelAction,
                color: Colors.WHITE,
              }}>
              Rejected
            </Text>
          </TouchableOpacity>
        </View>
      )}
      <ModalApproval
        isOpen={showApproval}
        onClose={() => setShowApproval(false)}
        onSubmit={(message: string) => onSubmitPost(message)}
        Op={operation}
      />
      <ModalAttachment
        isOpen={showAttachment}
        data={attachments}
        onClose={() => setShowAttachment(false)}
      />
    </View>
  );
};

export default container(Layout, false);

const styles = StyleSheet.create({
  header: {
    // flexDirection: 'row',
    // justifyContent: 'space-between',
  },
  headerItemContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  headerText: {
    ...Typography.textLg,
    color: Colors.BLACK,
  },
  section: {
    marginBottom: Mixins.scaleSize(10),
  },
  card: {
    marginBottom: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
  },
  cardHeader: {
    paddingHorizontal: Mixins.scaleSize(14),
    paddingVertical: Mixins.scaleSize(10),
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: Colors.SHADES.gray[200],
  },
  cardHeaderTitle: {
    ...Typography.textMdPlusSemiBold,
    color: Colors.SHADES.dark[700],
  },
  cardBody: {
    paddingHorizontal: Mixins.scaleSize(14),
    paddingVertical: Mixins.scaleSize(10),
    // flexDirection: 'column',
  },
  containerPrint: {
    position: 'relative',
    borderRadius: 100,
    borderColor: Colors.PRIMARY.green,
    borderWidth: 1,
    // width: Mixins.scaleSize(40),
    // height: Mixins.scaleSize(40),
    padding: Mixins.scaleSize(7),
    backgroundColor: Colors.WHITE,
  },
  buttonAction: {
    flex: 1,
    flexDirection: 'row',
    padding: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  labelAction: {
    ...Typography.textMdPlusSemiBold,
  },
});
