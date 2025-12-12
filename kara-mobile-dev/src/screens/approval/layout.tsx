/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {
  ActivityIndicator,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import s from '@components/styles/index';
import {Card, CheckBox, Divider} from '@ui-kitten/components';
import {Mixins, Typography, Colors} from 'utils';
import {
  Airdrop,
  Buildings2,
  ClipboardClose,
  ClipboardTick,
  GlobalSearch,
} from 'iconsax-react-native';
import {default as theme} from '../../custom-theme.json';
import {useActions} from '@overmind/index';
import {useIsFocused} from '@react-navigation/native';
import container from 'components/container';
import ApprovalItem from './card-item';
import ModalApproval from './modal-approval';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';
import DateTimePickerModal from 'react-native-modal-datetime-picker';
import ModalLoading from '@components/modal-loading';

import moment from 'moment';

type Props = {
  navigation: any;
};

const Layout = (props: Props) => {
  const isFocused = useIsFocused();
  // const ctx = React.useContext(ContainerContext);
  const {} = props;
  const actions = useActions();
  const [filter, setFilter] = React.useState({
    Start: new Date(moment().startOf('day')),
    End: new Date(moment().endOf('day')),
    GroupBy: 'Module',
    Status: 'PENDING',
  });
  const [showDateStart, setShowDateStart] = React.useState(false);
  const [showDateEnd, setShowDateEnd] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [loadingApproval, setLoadingAppoval] = React.useState(false);
  const [showApproval, setShowApproval] = React.useState(false);
  const [operation, setOperation] = React.useState('');
  const [data, setData] = React.useState([] as any[]);
  const [selectAll, setSelectAll] = React.useState(false);
  const tabs = [
    {key: 'PENDING', title: 'Need Approval'},
    {key: 'APPROVED', title: 'Approved'},
    {key: 'REJECTED', title: 'Rejected'},
  ];
  const init = () => {
    setLoading(true);
    actions
      .getsDataApproval(filter)
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
  }, [filter, isFocused]);

  const onSubmitPostGroup = (text: string) => {
    setShowApproval(false);
    setLoading(true);
    setLoadingAppoval(true);
    const filteredItem = data.filter(item => item.checked);
    const payload: any[] = filteredItem.map((item: any) => {
      return {
        ...item,
        Type: filter.GroupBy === 'Site' ? item.SiteID : item.Group,
        GroupBy: filter.GroupBy,
        Op: operation,
        Text: text,
      };
    });
    actions
      .postApprovalByGroup(payload)
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
  return (
    <View style={s.container}>
      <View style={{flex: 1}}>
        <ScrollView showsVerticalScrollIndicator={false}>
          <Card style={styles.card}>
            <View
              style={{...s.row, gap: 10, marginBottom: Mixins.scaleSize(10)}}>
              <View style={{flex: 1}}>
                <Text style={[styles.label]}>Date From</Text>
                <TouchableOpacity
                  onPress={() => setShowDateStart(true)}
                  style={{
                    borderColor: Colors.SHADES.gray[300],
                    borderRadius: Mixins.scaleSize(4),
                    borderWidth: 1,
                    paddingHorizontal: Mixins.scaleSize(10),
                    paddingVertical: Mixins.scaleSize(5),
                  }}>
                  <Text style={styles.value}>
                    {moment(filter.Start).format('DD/MMM/YYYY')}
                  </Text>
                </TouchableOpacity>
                <DateTimePickerModal
                  isVisible={showDateStart}
                  mode="date"
                  onConfirm={date => {
                    setFilter({
                      ...filter,
                      Start: new Date(moment(date).startOf('day')),
                    });
                    setShowDateStart(false);
                  }}
                  onCancel={() => setShowDateStart(false)}
                />
              </View>
              <View style={{flex: 1}}>
                <Text style={[styles.label]}>Date to</Text>
                {/* <Datepicker
                  date={filter.End}
                  onSelect={nextDate => {
                    setFilter({
                      ...filter,
                      End: nextDate,
                    });
                  }}
                /> */}
                <TouchableOpacity
                  onPress={() => setShowDateEnd(true)}
                  style={{
                    borderColor: Colors.SHADES.gray[300],
                    borderRadius: Mixins.scaleSize(4),
                    borderWidth: 1,
                    paddingHorizontal: Mixins.scaleSize(10),
                    paddingVertical: Mixins.scaleSize(5),
                  }}>
                  <Text style={styles.value}>
                    {moment(filter.End).format('DD/MMM/YYYY')}
                  </Text>
                </TouchableOpacity>
                <DateTimePickerModal
                  isVisible={showDateEnd}
                  mode="date"
                  onConfirm={date => {
                    setFilter({
                      ...filter,
                      End: new Date(moment(date).endOf('day')),
                    });
                    setShowDateEnd(false);
                  }}
                  onCancel={() => setShowDateEnd(false)}
                />
              </View>
            </View>
            <Divider />
            <View
              style={{
                ...s.row,
                justifyContent: 'space-around',
                marginTop: Mixins.scaleSize(10),
              }}>
              <TouchableOpacity
                onPress={() => setFilter({...filter, GroupBy: 'Module'})}
                style={{...styles.buttonGroup}}>
                <View
                  style={[
                    styles.iconContainer,
                    filter.GroupBy === 'Module'
                      ? {
                          borderWidth: 1,
                          borderColor: '#E1514A',
                          borderRadius: Mixins.scaleSize(6),
                        }
                      : {},
                  ]}>
                  <Buildings2
                    size={Mixins.scaleSize(30)}
                    color={
                      filter.GroupBy === 'Module'
                        ? '#E1514A'
                        : Colors.SHADES.gray[500]
                    }
                    variant="TwoTone"
                  />
                </View>
                <Text
                  style={{
                    ...Typography.textLg,
                    color:
                      filter.GroupBy === 'Module'
                        ? Colors.BLACK
                        : Colors.SHADES.gray[500],
                  }}>
                  Module
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={() => setFilter({...filter, GroupBy: 'Site'})}
                style={{...styles.buttonGroup}}>
                <View
                  style={[
                    styles.iconContainer,
                    filter.GroupBy === 'Site'
                      ? {
                          borderWidth: 1,
                          borderColor: '#E1514A',
                          borderRadius: Mixins.scaleSize(6),
                        }
                      : {},
                  ]}>
                  <GlobalSearch
                    size={Mixins.scaleSize(30)}
                    color={
                      filter.GroupBy === 'Site'
                        ? '#E1514A'
                        : Colors.SHADES.gray[500]
                    }
                    variant="TwoTone"
                  />
                </View>
                <Text
                  style={{
                    ...Typography.textLg,
                    color:
                      filter.GroupBy === 'Site'
                        ? Colors.BLACK
                        : Colors.SHADES.gray[500],
                  }}>
                  Site
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={() => setFilter({...filter, GroupBy: 'Object'})}
                style={{...styles.buttonGroup}}>
                <View
                  style={[
                    styles.iconContainer,
                    filter.GroupBy === 'Object'
                      ? {
                          borderWidth: 1,
                          borderColor: '#E1514A',
                          borderRadius: Mixins.scaleSize(6),
                        }
                      : {},
                  ]}>
                  <Airdrop
                    size={Mixins.scaleSize(30)}
                    color={
                      filter.GroupBy === 'Object'
                        ? '#E1514A'
                        : Colors.SHADES.gray[500]
                    }
                    variant="TwoTone"
                  />
                </View>
                <Text
                  style={{
                    ...Typography.textLg,
                    color:
                      filter.GroupBy === 'Object'
                        ? Colors.BLACK
                        : Colors.SHADES.gray[500],
                  }}>
                  Object
                </Text>
              </TouchableOpacity>
            </View>
          </Card>
          <View
            style={{
              ...s.row,
              borderColor: Colors.SHADES.gray[300],
              borderBottomWidth: 1,
              marginBottom: Mixins.scaleSize(10),
            }}>
            {tabs.map(item => (
              <TouchableOpacity
                onPress={() => setFilter({...filter, Status: item.key})}
                key={item.key}
                style={[
                  styles.buttonTab,
                  filter.Status === item.key
                    ? {
                        borderBottomWidth: 2,
                        borderColor: '#E1514A',
                      }
                    : {},
                ]}>
                <Text
                  style={{
                    ...styles.buttonTabText,
                    color:
                      filter.Status === item.key
                        ? '#E1514A'
                        : Colors.SHADES.dark[700],
                  }}>
                  {item.title}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
          {loading ? (
            <View>
              <ActivityIndicator color={Colors.PRIMARY.red} />
            </View>
          ) : (
            <View>
              {data.length === 0 ? (
                <View
                  style={{
                    flex: 1,
                    justifyContent: 'center',
                    paddingTop: Mixins.scaleSize(20),
                  }}>
                  <Text
                    style={{
                      ...Typography.textLgSemiBold,
                      textAlign: 'center',
                      color: Colors.BLACK,
                    }}>
                    No Data
                  </Text>
                </View>
              ) : (
                <View style={{marginBottom: Mixins.scaleSize(10)}}>
                  <View
                    style={{
                      ...s.row,
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      marginBottom: Mixins.scaleSize(10),
                    }}>
                    <Text
                      style={{
                        ...Typography.textMd,
                        color: Colors.SHADES.dark[700],
                      }}>
                      Data
                    </Text>
                    <View
                      style={{
                        ...s.row,
                        alignItems: 'center',
                        gap: 5,
                      }}>
                      <CheckBox
                        style={{}}
                        checked={selectAll}
                        onChange={nextChecked => {
                          setSelectAll(nextChecked);
                          const newData = data.map((item: any) => {
                            return {...item, checked: nextChecked};
                          });
                          setData([...newData]);
                        }}>
                        {evaProps => (
                          <Text
                            {...evaProps}
                            style={{
                              marginLeft: Mixins.scaleSize(5),
                              ...Typography.textMd,
                              color: Colors.SHADES.dark[700],
                            }}>
                            Select All
                          </Text>
                        )}
                      </CheckBox>
                    </View>
                  </View>
                  <View style={{flexDirection: 'column', gap: 10}}>
                    {data.map((item, i) => (
                      <ApprovalItem
                        key={i}
                        item={item}
                        loading={loading}
                        onSelect={() =>
                          props.navigation.navigate('ApprovalDetail', {
                            payload: {
                              Date: item.Date,
                              GroupBy: filter.GroupBy,
                              Status: filter.Status,
                              Type:
                                filter.GroupBy === 'Site'
                                  ? item.SiteID
                                  : item.Group,
                              Total: item.Total,
                            },
                            title: item.Group,
                          })
                        }
                        onChecked={nextValue => {
                          let newData: any = [...data];
                          newData[i] = {...newData[i], checked: nextValue};
                          setData([...newData]);
                        }}
                      />
                    ))}
                  </View>
                </View>
              )}
            </View>
          )}
        </ScrollView>
      </View>
      {filter.Status === 'PENDING' && (
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
        onSubmit={(message: string) => onSubmitPostGroup(message)}
        Op={operation}
      />
      <ModalLoading show={loadingApproval} />
    </View>
  );
};

export default container(Layout, false);

const styles = StyleSheet.create({
  card: {
    borderRadius: Mixins.scaleSize(8),
    padding: 0,
    marginBottom: Mixins.scaleSize(10),
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  value: {
    ...Typography.textLg,
    color: Colors.SHADES.dark[700],
  },
  buttonGroup: {
    padding: Mixins.scaleSize(10),
    textAlign: 'center',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: Mixins.scaleSize(8),
  },
  iconContainer: {
    marginBottom: Mixins.scaleSize(5),
    padding: Mixins.scaleSize(10),
  },
  buttonTab: {
    flex: 1,
    justifyContent: 'center',
    paddingVertical: Mixins.scaleSize(10),
  },
  buttonTabText: {
    ...Typography.textMdPlusSemiBold,
    textAlign: 'center',
    color: Colors.SHADES.dark[700],
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
