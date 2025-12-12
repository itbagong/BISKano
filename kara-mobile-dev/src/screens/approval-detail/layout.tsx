/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {useActions} from '@overmind/index';
import {useIsFocused} from '@react-navigation/native';
import s from '@components/styles';
import {Colors, Helper, Mixins, Typography} from 'utils';
import {ClipboardClose, ClipboardTick} from 'iconsax-react-native';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';
import {default as theme} from '../../custom-theme.json';
import moment from 'moment';
import Card from 'components/card';
import container from 'components/container';
import {CheckBox, Icon} from '@ui-kitten/components/ui';
import ModalApproval from '../approval/modal-approval';
import ModalAttachment from './modal-attachment';
import ModalLoading from 'components/modal-loading';

type Props = {
  route: any;
  navigation: any;
};

const Layout = (props: Props) => {
  const {route, navigation} = props;
  const isFocused = useIsFocused();
  const {getsDataApprovalJournal, postApproval, getsAssetByJournal} =
    useActions();
  const [data, setData] = React.useState([] as any[]);
  const [loading, setLoading] = React.useState(false);
  const [showApproval, setShowApproval] = React.useState(false);
  const [operation, setOperation] = React.useState('');
  const [loadingApproval, setLoadingAppoval] = React.useState(false);

  const capitalizeFirstLetter = (string: string) => {
    return string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();
  };

  React.useLayoutEffect(() => {
    navigation.setOptions({
      headerTitle: 'Detail ' + capitalizeFirstLetter(route.params.title),
    });

    return () => {};
  }, [isFocused]);

  const init = () => {
    setLoading(true);
    getsDataApprovalJournal(route.params.payload)
      .then(res => {
        setData(res);
      })
      .finally(() => setLoading(false));
  };
  React.useEffect(() => {
    if (isFocused) {
      init();
    }
    return () => {};
  }, [isFocused, route]);

  const onSubmitPost = (text: string) => {
    setShowApproval(false);
    setLoading(true);
    setLoadingAppoval(true);
    const filteredItem = data.filter(item => item.checked);
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
      post.push(postApproval({type: 'fico', payload: payloadFico}));
    }
    if (payloadSCM.length > 0) {
      post.push(postApproval({type: 'scm', payload: payloadSCM}));
    }
    if (payloadMFG.length > 0) {
      post.push(postApproval({type: 'mfg', payload: payloadMFG}));
    }
    Promise.all(post)
      .then(() => {
        Toast.show({
          type: ALERT_TYPE.SUCCESS,
          title: 'Success',
          textBody: `Data has been ${
            operation === 'Approve' ? 'Approved' : 'Rejected'
          }`,
        });
        init();
      })
      .finally(() => {
        setLoading(false);
        setLoadingAppoval(false);
      });
  };
  const onOpenApproval = (op: string) => {
    if (!data.find(o => o.checked)) {
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody: 'Opps, No data is selected!',
      });
    }
    setOperation(op);
    setTimeout(() => {
      setShowApproval(true);
    }, 100);
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
  return (
    <View style={s.container}>
      <View style={{flex: 1}}>
        <ScrollView showsVerticalScrollIndicator={false}>
          <View style={styles.header}>
            <Text
              style={{
                ...Typography.textLgPlus,
                color: Colors.SHADES.dark[500],
              }}>
              {route.params.payload.Type}
            </Text>
            {/* <Text
              style={{
                ...Typography.textLgSemiBold,
                color: Colors.SHADES.dark[700],
              }}>
              {moment(route.params.payload.Date).format('DD - MMMM - YYYY')}
            </Text> */}
            <Text
              style={{
                ...Typography.headerSmSemiBold,
                color: Colors.SHADES.green[500],
              }}>
              Total Amount:{' '}
              {Helper.currencyFormat(route.params.payload.Total, '')}
            </Text>
          </View>
          {data.map((item: any, i) => (
            <Card key={i} containerStyle={styles.card}>
              <View style={styles.cardHeader}>
                <View style={{...s.row, alignItems: 'center', gap: 10}}>
                  <CheckBox
                    style={{}}
                    checked={item.checked}
                    onChange={nextChecked => {
                      let newData: any = [...data];
                      newData[i] = {...newData[i], checked: nextChecked};
                      setData([...newData]);
                    }}
                  />
                  <Text
                    style={{
                      ...Typography.textMdPlusSemiBold,
                      color: Colors.SHADES.dark[700],
                    }}>
                    {item.AccountID}
                  </Text>
                </View>
                <Text
                  style={{
                    ...Typography.textMdPlusSemiBold,
                    color: Colors.SHADES.green[500],
                  }}>
                  {Helper.currencyFormat(item.Amount)}
                </Text>
              </View>
              <View style={styles.cardBody}>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    Object
                  </Text>
                  <Text
                    style={{
                      flex: 1.5,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    : {item.SourceType}
                  </Text>
                </View>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    Date
                  </Text>
                  <Text
                    style={{
                      flex: 1.5,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    : {moment(item.TrxDate).format('DD MMMM YYYY hh:mm')}
                  </Text>
                </View>
                <View style={s.row}>
                  <Text
                    style={{
                      flex: 1,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    Text
                  </Text>
                  <Text
                    style={{
                      flex: 1.5,
                      ...Typography.textMdPlus,
                      color: Colors.BLACK,
                    }}>
                    : {item.Text}
                  </Text>
                </View>
                <View
                  style={{
                    ...s.row,
                    marginTop: Mixins.scaleSize(10),
                    gap: Mixins.scaleSize(10),
                  }}>
                  <TouchableOpacity
                    onPress={() => {
                      onShowAttachment(item);
                    }}
                    style={{
                      ...styles.buttonAction,
                      flex: 1,
                      borderColor: Colors.SHADES.green[500],
                      backgroundColor: Colors.SHADES.green[50],
                      borderWidth: 1,
                    }}>
                    <Icon
                      fill={Colors.SHADES.green[500]}
                      name="attach-outline"
                      style={{
                        width: Mixins.scaleSize(20),
                        height: Mixins.scaleSize(20),
                      }}
                    />
                    <Text
                      style={{
                        ...styles.labelAction,
                        color: Colors.SHADES.green[500],
                      }}>
                      Attachment
                    </Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    onPress={() => {
                      navigation.navigate('ApprovalPreview', {
                        SourceType: item.SourceType,
                        SourceJournalID: item.SourceID,
                        Item: item,
                        Status: route.params.payload.Status,
                        Name: 'Default',
                        VoucherNo: '',
                      });
                    }}
                    style={{
                      ...styles.buttonAction,
                      flex: 1,
                      borderColor: Colors.SHADES.green[500],
                      backgroundColor: Colors.SHADES.green[50],
                      borderWidth: 1,
                    }}>
                    <Icon
                      fill={Colors.SHADES.green[500]}
                      name="eye-outline"
                      style={{
                        width: Mixins.scaleSize(20),
                        height: Mixins.scaleSize(20),
                      }}
                    />
                    <Text
                      style={{
                        ...styles.labelAction,
                        color: Colors.SHADES.green[500],
                      }}>
                      Preview
                    </Text>
                  </TouchableOpacity>
                </View>
              </View>
            </Card>
          ))}
        </ScrollView>
      </View>
      {route.params.payload.Status === 'PENDING' && (
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
      <ModalLoading show={loadingApproval} />

    </View>
  );
};

export default container(Layout, false);

const styles = StyleSheet.create({
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
  header: {
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
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
  cardBody: {
    paddingHorizontal: Mixins.scaleSize(14),
    paddingVertical: Mixins.scaleSize(10),
    flexDirection: 'column',
    alignItems: 'center',
  },
});
