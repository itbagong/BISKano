/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {useActions} from '@overmind/index';
import s from '@components/styles/index';
import {Card, Icon} from '@ui-kitten/components/ui';
import {Colors, Mixins, Typography} from '@utils/index';
import moment from 'moment';
import AccordionItem from '@components/accordion-item';
import {default as theme} from '../../custom-theme.json';
import container, {ContainerContext} from 'components/container';
import {useIsFocused} from '@react-navigation/native';

type Props = {
  navigation: any;
  route: any;
};

const Layout = (props: Props) => {
  const {route, navigation} = props;
  const isFocused = useIsFocused();
  const ctx = React.useContext(ContainerContext);
  const {getRoutine, getsRoutineDetails, getsRoutineChecklist} = useActions();
  const [data, setData] = React.useState({} as any);
  const [dataDetails, setDataDetails] = React.useState([]);
  const [loading, setLoading] = React.useState(false);

  const init = () => {
    getRoutine({_id: route?.params._id, Sort: ['-_id']}).then((res: any) => {
      if (res.count > 0) {
        setData(res.data[0]);
      }
    });
    setLoading(true);
    getsRoutineDetails({_id: route?.params._id, Sort: ['-_id']})
      .then((res: any) => {
        setDataDetails(res.data);
      })
      .finally(() => setLoading(false));
  };
  React.useLayoutEffect(() => {
    ctx.setRefreshCallback({
      func: async () => {
        init();
      },
    });
    return () => {};
  }, [isFocused]);
  React.useEffect(() => {
    init();
    return () => {};
  }, [isFocused]);

  const renderStatus = (status: string) => {
    let text = '';
    let bgColor = 'black';
    let txtColor = 'black';
    switch (status) {
      case 'NotCheckedYet':
        text = 'Not Checked Yet';
        bgColor = Colors.SHADES.gray[200];
        break;
      case 'NeedRepair':
        text = 'Need Repair';
        bgColor = Colors.SHADES.red[200];
        txtColor = 'white';
        break;
      case 'RunningWell':
        text = 'Running Well';
        bgColor = theme['color-success-200'];
        txtColor = 'white';
        break;
      default:
        break;
    }
    return (
      <View style={{flex: 1, flexDirection: 'row'}}>
        <Text style={{...Typography.textMdPlusSemiBold}}>: </Text>
        <Text
          style={{
            backgroundColor: bgColor,
            paddingHorizontal: Mixins.scaleSize(10),
            borderRadius: 4,
            ...Typography.textMdPlusSemiBold,
            color: txtColor,
          }}>
          {text}
        </Text>
      </View>
    );
  };
  return (
    <View style={s.container}>
      <Card style={styles.card}>
        <View style={s.row}>
          <View
            style={{
              flex: 1,
              borderRightWidth: 1,
              borderColor: Colors.SHADES.gray[500],
              justifyContent: 'center',
              alignItems: 'center',
            }}>
            <Text style={{...Typography.textLgSemiBold, color: Colors.BLACK}}>
              {data?.SiteName}
            </Text>
            <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
              Site
            </Text>
          </View>
          <View
            style={{flex: 1, justifyContent: 'center', alignItems: 'center'}}>
            <Text style={{...Typography.textLgSemiBold, color: Colors.BLACK}}>
              {moment(data?.ExecutionDate).format('DD MMM YYYY')}
            </Text>
            <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
              Date
            </Text>
          </View>
        </View>
      </Card>
      {loading ? (
        <View
          style={{flex: 1, alignContent: 'center', justifyContent: 'center'}}>
          <ActivityIndicator color={Colors.PRIMARY.red} />
        </View>
      ) : (
        <View style={{}}>
          {dataDetails.map((item: any, i: number) => (
            <AccordionItem key={i} title={item._id}>
              <View style={styles.rowForm}>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                  }}>
                  Asset Name
                </Text>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlusSemiBold,
                    color: Colors.BLACK,
                  }}>
                  : {item.AssetName}
                </Text>
              </View>
              <View style={styles.rowForm}>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                  }}>
                  Type Name
                </Text>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlusSemiBold,
                    color: Colors.BLACK,
                  }}>
                  : {item.AssetTypeName}
                </Text>
              </View>
              <View style={styles.rowForm}>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                  }}>
                  Drive Type
                </Text>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlusSemiBold,
                    color: Colors.BLACK,
                  }}>
                  : {item.DriveType}
                </Text>
              </View>
              <View style={styles.rowForm}>
                <Text
                  style={{
                    flex: 1,
                    ...Typography.textMdPlus,
                    color: Colors.BLACK,
                  }}>
                  Status Condition
                </Text>
                {renderStatus(item.StatusCondition)}
              </View>
              <View
                style={{
                  ...styles.rowForm,
                  gap: Mixins.scaleSize(10),
                  marginTop: Mixins.scaleSize(10),
                }}>
                <TouchableOpacity
                  onPress={() => {
                    getsRoutineChecklist({RoutineDetailID: item._id}).then(
                      () => {
                        navigation.navigate('RoutineAssetDetails', {
                          routineAsset: item,
                          routine: data,
                          isReadOnly: route?.params?.isReadOnly,
                        });
                      },
                    );
                  }}
                  style={{
                    ...styles.buttonAction,
                    backgroundColor: theme['color-success-300'],
                  }}>
                  <Icon
                    fill={theme['color-success-900']}
                    name="eye-outline"
                    style={{
                      width: Mixins.scaleSize(20),
                      height: Mixins.scaleSize(20),
                    }}
                  />
                  <Text
                    style={{
                      ...styles.labelAction,
                      color: theme['color-success-900'],
                    }}>
                    View
                  </Text>
                </TouchableOpacity>
              </View>
            </AccordionItem>
          ))}
        </View>
      )}
    </View>
  );
};

export default container(Layout);

const styles = StyleSheet.create({
  card: {
    borderRadius: Mixins.scaleSize(12),
    marginBottom: Mixins.scaleSize(10),
  },
  rowForm: {
    flexDirection: 'row',
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
